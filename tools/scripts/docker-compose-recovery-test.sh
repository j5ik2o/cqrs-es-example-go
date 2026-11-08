#!/usr/bin/env bash

# Exercise the RMU recovery path locally. The recovery worker uses the same
# handler as the Lambda function: the startup pass models an SQS OnFailure
# notification and the interval models the EventBridge schedule. The script
# simulates a lost projection, records the durable recovery task a permanent
# failure would leave, and checks that the worker restores the read model
# without duplicating rows.

set -eu

cd "$(dirname "$0")" || exit

export AWS_REGION=${AWS_REGION:-ap-northeast-1}

F_OPTION="-f ../docker-compose/docker-compose-applications.yml -f ../docker-compose/docker-compose-e2e-test.yml"

compose() {
  docker compose -p cqrs-es-example-go -f ../docker-compose/docker-compose-databases.yml ${F_OPTION} "$@"
}

mysql_query() {
  compose exec -T mysql mysql --protocol=TCP -h127.0.0.1 -uceer -pceer -N -B ceer -e "$1" 2>/dev/null
}

fail() {
  echo "Error: $*" >&2
  exit 1
}

echo "Verifying the RMU recovery path"

aggregate_id=$(mysql_query "SELECT id FROM group_chats WHERE name = 'group-chat-after-restart' LIMIT 1")
[ -n "${aggregate_id}" ] || fail "the group chat created after the RMU restart was not found"
case "${aggregate_id}" in
*[!A-Za-z0-9_-]*) fail "unexpected aggregate id: ${aggregate_id}" ;;
esac

seq_nr=$(mysql_query "SELECT seq_nr FROM read_model_positions WHERE aggregate_id = '${aggregate_id}'")
[ -n "${seq_nr}" ] || fail "no read model position for ${aggregate_id}"

messages_before=$(mysql_query "SELECT COUNT(*) FROM messages WHERE group_chat_id = '${aggregate_id}'")
members_before=$(mysql_query "SELECT COUNT(*) FROM members WHERE group_chat_id = '${aggregate_id}'")
group_chats_before=$(mysql_query "SELECT COUNT(*) FROM group_chats WHERE id = '${aggregate_id}'")
[ "${group_chats_before}" = "1" ] || fail "expected one group chat row before recovery"

store_id=$(compose exec -T localstack awslocal dynamodb get-item \
  --table-name head \
  --key '{"aid":{"S":"__config__"}}' \
  --query 'Item.store_id.S' \
  --output text 2>/dev/null)
[ -n "${store_id}" ] && [ "${store_id}" != "None" ] || fail "the head table has no store_id configuration item"

echo "Recovery fixture: aggregate=${aggregate_id} seq_nr=${seq_nr} store=${store_id}"

# Stop the stream reader and the worker so the simulated loss is deterministic.
compose stop read-model-updater-1 recovery-worker >/dev/null

# Simulate a projection that lost this aggregate: remove the read model rows
# and the position, then record the durable recovery task a permanent failure
# would leave behind.
mysql_query "DELETE FROM messages WHERE group_chat_id = '${aggregate_id}'"
mysql_query "DELETE FROM members WHERE group_chat_id = '${aggregate_id}'"
mysql_query "DELETE FROM group_chats WHERE id = '${aggregate_id}'"
mysql_query "DELETE FROM read_model_positions WHERE aggregate_id = '${aggregate_id}'"
now=$(date '+%Y-%m-%d %H:%M:%S')
mysql_query "INSERT INTO read_model_recovery_tasks (store_id, projection_name, projection_version, aggregate_id, seq_nr, failure_type, error_message, status, created_at, updated_at) VALUES ('${store_id}', 'group-chat-read-model', '1', '${aggregate_id}', ${seq_nr}, 'journal_gap', 'simulated permanent failure for local verification', 'pending', '${now}', '${now}')"

# Start the worker. Its startup pass is the SQS OnFailure equivalent: it
# reconciles every head from the journal and then replays pending tasks.
# --no-deps is required: a plain start would rerun the one-shot
# dynamodb-setup job, which fails because the tables already exist.
compose up -d --no-deps recovery-worker >/dev/null

status=""
for attempt in $(seq 1 60); do
  status=$(mysql_query "SELECT status FROM read_model_recovery_tasks WHERE store_id = '${store_id}' AND projection_name = 'group-chat-read-model' AND aggregate_id = '${aggregate_id}' AND seq_nr = ${seq_nr}")
  [ "${status}" = "completed" ] && break
  sleep 1
done
if [ "${status}" != "completed" ]; then
  compose logs --no-color --tail=100 recovery-worker
  fail "the recovery worker did not complete the pending task"
fi

messages_after=$(mysql_query "SELECT COUNT(*) FROM messages WHERE group_chat_id = '${aggregate_id}'")
members_after=$(mysql_query "SELECT COUNT(*) FROM members WHERE group_chat_id = '${aggregate_id}'")
group_chats_after=$(mysql_query "SELECT COUNT(*) FROM group_chats WHERE id = '${aggregate_id}'")
position_after=$(mysql_query "SELECT seq_nr FROM read_model_positions WHERE aggregate_id = '${aggregate_id}'")
pending_after=$(mysql_query "SELECT COUNT(*) FROM read_model_recovery_tasks WHERE store_id = '${store_id}' AND projection_name = 'group-chat-read-model' AND status = 'pending'")

[ "${messages_after}" = "${messages_before}" ] ||
  fail "messages count changed: before=${messages_before} after=${messages_after}"
[ "${members_after}" = "${members_before}" ] ||
  fail "members count changed: before=${members_before} after=${members_after}"
[ "${group_chats_after}" = "${group_chats_before}" ] ||
  fail "group_chats count changed: before=${group_chats_before} after=${group_chats_after}"
[ "${position_after}" = "${seq_nr}" ] ||
  fail "position after recovery is ${position_after}, expected ${seq_nr}"
[ "${pending_after}" = "0" ] ||
  fail "${pending_after} recovery tasks are still pending"

echo "Recovery worker restored the read model without duplicates"

# The interval pass is the EventBridge schedule safety net. It must keep
# running even when no SQS notification arrives.
before_lines=$(compose logs --no-color recovery-worker | grep -c 'recovery pass finished' || true)
[ "${before_lines}" -ge 1 ] || fail "the recovery worker has not logged a completion line"
after_lines=${before_lines}
for attempt in $(seq 1 30); do
  after_lines=$(compose logs --no-color recovery-worker | grep -c 'recovery pass finished' || true)
  [ "${after_lines}" -gt "${before_lines}" ] && break
  sleep 1
done
[ "${after_lines}" -gt "${before_lines}" ] ||
  fail "the scheduled recovery pass did not run (lines before=${before_lines} after=${after_lines})"

last_line=$(compose logs --no-color recovery-worker | grep 'recovery pass finished' | tail -1)
grep -q '"pending":0' <<<"${last_line}" || fail "the scheduled pass left pending tasks: ${last_line}"
grep -q '"resync_failed":0' <<<"${last_line}" || fail "the scheduled pass reported a resync failure: ${last_line}"
echo "Scheduled replay is running (${before_lines} -> ${after_lines} completion lines)"

# The manual one-shot command must be idempotent as well. Applying the same
# tasks again may not duplicate rows.
compose run -T --rm --no-deps recovery-worker recoverRmu
messages_manual=$(mysql_query "SELECT COUNT(*) FROM messages WHERE group_chat_id = '${aggregate_id}'")
members_manual=$(mysql_query "SELECT COUNT(*) FROM members WHERE group_chat_id = '${aggregate_id}'")
group_chats_manual=$(mysql_query "SELECT COUNT(*) FROM group_chats WHERE id = '${aggregate_id}'")
[ "${messages_manual}" = "${messages_before}" ] || fail "recoverRmu duplicated message rows"
[ "${members_manual}" = "${members_before}" ] || fail "recoverRmu duplicated member rows"
[ "${group_chats_manual}" = "${group_chats_before}" ] || fail "recoverRmu duplicated group chat rows"
echo "Manual recoverRmu is idempotent"

compose up -d --no-deps read-model-updater-1 >/dev/null
echo "Recovery verification passed"
