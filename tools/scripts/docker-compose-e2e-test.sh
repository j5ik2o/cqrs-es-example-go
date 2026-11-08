#!/usr/bin/env bash

set -eu

# shellcheck disable=SC2046
cd $(dirname "$0") || exit

export ARCH=$(uname -m)
echo "ARCH=${ARCH}"

export AWS_REGION=${AWS_REGION:-ap-northeast-1}

F_OPTION="-f ../docker-compose/docker-compose-applications.yml -f ../docker-compose/docker-compose-e2e-test.yml"

compose() {
  docker compose -p cqrs-es-example-go -f ../docker-compose/docker-compose-databases.yml ${F_OPTION} "$@"
}

compose down -v --remove-orphans
compose up --remove-orphans --force-recreate --renew-anon-volumes -d "$@"

echo "Waiting for the e2e-test container to finish"
e2e_status=0
compose wait e2e-test || e2e_status=$?
if [ "${e2e_status}" -ne 0 ]; then
  echo "e2e-test failed with exit code ${e2e_status}"
  compose logs --no-color --tail=200 e2e-test
  exit "${e2e_status}"
fi
echo "e2e-test passed"

# The read model updater replays the head stream from TRIM_HORIZON after a
# restart. The services stay up so LocalStack keeps its in-memory tables, and
# only the updater restarts.
mysql_query() {
  compose exec -T mysql mysql --protocol=TCP -h127.0.0.1 -uceer -pceer -N -B ceer -e "$1" 2>/dev/null
}

positions_before=$(mysql_query "SELECT aggregate_id, seq_nr FROM read_model_positions ORDER BY aggregate_id")
messages_before=$(mysql_query "SELECT COUNT(*) FROM messages")
group_chats_before=$(mysql_query "SELECT COUNT(*) FROM group_chats")
if [ -z "${positions_before}" ]; then
  echo "Error: read_model_positions is empty before restart" >&2
  exit 1
fi

echo "Restarting read-model-updater-1 to exercise replay from committed positions"
compose restart read-model-updater-1

echo "Running the post-restart verification"
if ! compose run -T --rm --no-deps --entrypoint ./verify-after-restart.sh e2e-test; then
  echo "Post-restart verification failed"
  compose logs --no-color --tail=200 read-model-updater-1
  exit 1
fi

positions_after=$(mysql_query "SELECT aggregate_id, seq_nr FROM read_model_positions ORDER BY aggregate_id")
messages_after=$(mysql_query "SELECT COUNT(*) FROM messages")
group_chats_after=$(mysql_query "SELECT COUNT(*) FROM group_chats")

if [ "${messages_after}" -ne $((messages_before + 1)) ]; then
  echo "Error: messages count after restart is ${messages_after}, expected $((messages_before + 1))" >&2
  exit 1
fi
if [ "${group_chats_after}" -ne $((group_chats_before + 1)) ]; then
  echo "Error: group_chats count after restart is ${group_chats_after}, expected $((group_chats_before + 1))" >&2
  exit 1
fi

# Every aggregate that existed before the restart keeps its position: the
# replayed records are dropped as duplicates instead of being applied again.
while read -r aggregate_id seq_nr; do
  [ -z "${aggregate_id}" ] && continue
  after_seq_nr=$(echo "${positions_after}" | awk -v aggregate_id="${aggregate_id}" '$1 == aggregate_id { print $2 }')
  if [ "${after_seq_nr}" != "${seq_nr}" ]; then
    echo "Error: read_model_positions for ${aggregate_id} changed from ${seq_nr} to ${after_seq_nr} after restart" >&2
    exit 1
  fi
done <<< "${positions_before}"

echo "Post-restart replay verification passed"

echo "Running the RMU recovery worker verification"
if ! ./docker-compose-recovery-test.sh; then
  echo "Recovery verification failed"
  compose logs --no-color --tail=200 recovery-worker
  exit 1
fi

echo "Running the deploy asset verification against LocalStack"
if ! compose run -T --rm --no-deps --entrypoint ./verify-deploy-assets.sh e2e-test; then
  echo "Deploy asset verification failed"
  compose logs --no-color --tail=200 localstack
  exit 1
fi
