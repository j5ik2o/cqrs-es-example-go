#!/usr/bin/env bash

# Verify the deploy/rmu assets as far as a local environment allows:
#   - JSON syntax and shell syntax
#   - apply.sh and apply-alarms.sh dry runs (they must make no AWS calls)
#   - the queue attributes, schedule rule, metric filters, and alarms against
#     LocalStack through the AWS CLI
#
# The Lambda event source mapping, SNS notifications, and real AWS resources
# are out of scope. They need a deployed function and account that this task
# must not create.

set -euo pipefail

CONFIG_DIR=${DEPLOY_CONFIG_DIR:-/deploy/rmu}
ENDPOINT_URL=${AWS_ENDPOINT_URL:-http://localstack:4566}
REGION=${AWS_REGION:-ap-northeast-1}

fail() {
  echo "Error: $*" >&2
  exit 1
}

command -v aws >/dev/null 2>&1 || fail "aws CLI is required"
command -v jq >/dev/null 2>&1 || fail "jq is required"

for file in event-source-mapping.json cloudwatch-alarms.json; do
  jq empty "${CONFIG_DIR}/${file}" || fail "${file} is not valid JSON"
  echo "JSON OK: ${file}"
done

bash -n "${CONFIG_DIR}/apply.sh" || fail "apply.sh has a shell syntax error"
bash -n "${CONFIG_DIR}/apply-alarms.sh" || fail "apply-alarms.sh has a shell syntax error"
echo "Shell syntax OK: apply.sh apply-alarms.sh"

head_stream_arn="arn:aws:dynamodb:ap-northeast-1:000000000000:table/head/stream/2026-01-01T00:00:00.000"
plan_queue_url="http://localstack:4566/000000000000/cqrs-es-example-go-rmu-failure"

# Dry runs must not call AWS. Point the CLI at a closed port so an accidental
# call fails the command instead of reaching LocalStack.
plan=$(AWS_ENDPOINT_URL=http://127.0.0.1:1 "${CONFIG_DIR}/apply.sh" \
  --rmu-function test-rmu \
  --head-stream-arn "${head_stream_arn}" \
  --failure-queue-url "${plan_queue_url}" \
  --recovery-function test-recovery)
grep -q "dry-run: no AWS calls were made" <<<"${plan}" || fail "apply.sh did not report a dry run"
grep -q "batch 100, bisect true, ReportBatchItemFailures, retries 3, max age 86400s" <<<"${plan}" ||
  fail "apply.sh plan does not match the safety settings"
grep -q "retention 1209600s, visibility 960s" <<<"${plan}" ||
  fail "apply.sh plan does not match the queue settings"
grep -q "recoverRmuWorker" <<<"${plan}" || fail "apply.sh plan does not select the recovery worker"
grep -q "rate(5 minutes), ENABLED" <<<"${plan}" || fail "apply.sh plan does not match the schedule"
echo "Dry run OK: apply.sh"

alarm_plan=$(AWS_ENDPOINT_URL=http://127.0.0.1:1 "${CONFIG_DIR}/apply-alarms.sh" \
  --rmu-function test-rmu \
  --recovery-function test-recovery \
  --failure-queue-url "${plan_queue_url}")
grep -q "dry-run: no AWS calls were made" <<<"${alarm_plan}" || fail "apply-alarms.sh did not report a dry run"
filter_count=$(grep -c '^filter:' <<<"${alarm_plan}")
alarm_count=$(grep -c '^alarm:' <<<"${alarm_plan}")
[ "${filter_count}" -eq 4 ] || fail "expected 4 metric filters in the plan, got ${filter_count}"
[ "${alarm_count}" -eq 7 ] || fail "expected 7 alarms in the plan, got ${alarm_count}"
echo "Dry run OK: apply-alarms.sh (${filter_count} filters, ${alarm_count} alarms)"

export AWS_ENDPOINT_URL="${ENDPOINT_URL}"
export AWS_DEFAULT_REGION="${REGION}"
export AWS_REGION="${REGION}"

echo "Waiting for LocalStack at ${AWS_ENDPOINT_URL}"
for attempt in $(seq 1 60); do
  if aws dynamodb list-tables >/dev/null 2>&1 &&
    aws sqs list-queues >/dev/null 2>&1 &&
    aws events list-rules >/dev/null 2>&1 &&
    aws logs describe-log-groups >/dev/null 2>&1 &&
    aws cloudwatch describe-alarms >/dev/null 2>&1; then
    break
  fi
  if [ "${attempt}" -eq 60 ]; then
    fail "LocalStack did not become ready"
  fi
  sleep 2
done
echo "LocalStack is ready"

# Apply the failure queue settings exactly as apply.sh does, then read them
# back. This proves the AWS CLI command shape and values work against a real
# SQS API implementation (LocalStack).
retention=$(jq -r '.failureQueue.messageRetentionPeriodSeconds' "${CONFIG_DIR}/event-source-mapping.json")
visibility=$(jq -r '.failureQueue.visibilityTimeoutSeconds' "${CONFIG_DIR}/event-source-mapping.json")
queue_url=$(aws sqs create-queue \
  --queue-name cqrs-es-example-go-rmu-failure \
  --query 'QueueUrl' \
  --output text)
aws sqs set-queue-attributes \
  --queue-url "${queue_url}" \
  --attributes "{\"MessageRetentionPeriod\":\"${retention}\",\"VisibilityTimeout\":\"${visibility}\"}"
queue_attributes=$(aws sqs get-queue-attributes \
  --queue-url "${queue_url}" \
  --attribute-names All \
  --output json)
[ "$(jq -r '.Attributes.MessageRetentionPeriod' <<<"${queue_attributes}")" = "${retention}" ] ||
  fail "queue MessageRetentionPeriod does not match the config"
[ "$(jq -r '.Attributes.VisibilityTimeout' <<<"${queue_attributes}")" = "${visibility}" ] ||
  fail "queue VisibilityTimeout does not match the config"
echo "SQS OK: retention=${retention}s visibility=${visibility}s"

# Put the schedule rule with the config values. The target is a Lambda ARN in
# production; LocalStack cannot verify that without a deployed function, so
# only the rule itself is checked.
schedule_rule=$(jq -r '.recoveryWorkerSchedule.ruleName' "${CONFIG_DIR}/event-source-mapping.json")
schedule_expression=$(jq -r '.recoveryWorkerSchedule.scheduleExpression' "${CONFIG_DIR}/event-source-mapping.json")
schedule_state=$(jq -r '.recoveryWorkerSchedule.state' "${CONFIG_DIR}/event-source-mapping.json")
aws events put-rule \
  --name "${schedule_rule}" \
  --schedule-expression "${schedule_expression}" \
  --state "${schedule_state}" >/dev/null
schedule_output=$(aws events describe-rule --name "${schedule_rule}" --output json)
[ "$(jq -r '.ScheduleExpression' <<<"${schedule_output}")" = "${schedule_expression}" ] ||
  fail "schedule expression does not match the config"
[ "$(jq -r '.State' <<<"${schedule_output}")" = "${schedule_state}" ] ||
  fail "schedule state does not match the config"
echo "EventBridge OK: ${schedule_rule} (${schedule_expression}, ${schedule_state})"

# apply-alarms.sh only touches Logs and CloudWatch, so it can run against
# LocalStack end to end. Alarm actions are omitted on purpose: this task must
# not create an SNS topic.
"${CONFIG_DIR}/apply-alarms.sh" \
  --rmu-function test-rmu \
  --recovery-function test-recovery \
  --failure-queue-url "${queue_url}" \
  --apply

expected_filters=$(jq '.metricFilters | length' "${CONFIG_DIR}/cloudwatch-alarms.json")
actual_filters=$(aws logs describe-metric-filters \
  --log-group-name /aws/lambda/test-recovery \
  --query 'length(metricFilters)' \
  --output text)
[ "${actual_filters}" = "${expected_filters}" ] ||
  fail "expected ${expected_filters} metric filters, got ${actual_filters}"

expected_alarms=$(jq '.alarms | length' "${CONFIG_DIR}/cloudwatch-alarms.json")
actual_alarms=$(aws cloudwatch describe-alarms \
  --alarm-name-prefix rmu- \
  --query 'length(MetricAlarms)' \
  --output text)
[ "${actual_alarms}" = "${expected_alarms}" ] ||
  fail "expected ${expected_alarms} alarms, got ${actual_alarms}"

pending_eval=$(aws cloudwatch describe-alarms \
  --alarm-names rmu-recovery-pending-tasks \
  --query 'MetricAlarms[0].EvaluationPeriods' \
  --output text)
expected_eval=$(jq -r '.alarms[] | select(.name == "rmu-recovery-pending-tasks") | .evaluationPeriods' \
  "${CONFIG_DIR}/cloudwatch-alarms.json")
[ "${pending_eval}" = "${expected_eval}" ] ||
  fail "alarm rmu-recovery-pending-tasks has EvaluationPeriods ${pending_eval}, expected ${expected_eval}"

echo "CloudWatch OK: ${actual_filters} metric filters, ${actual_alarms} alarms"
echo "Deploy asset verification passed"
