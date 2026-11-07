#!/usr/bin/env bash
#
# Apply the RMU event source mapping settings from event-source-mapping.json.
# The script is a dry run unless --apply is passed. Read-only lookups are
# skipped in a dry run, so a plan needs no AWS credentials.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CONFIG="${SCRIPT_DIR}/event-source-mapping.json"

APPLY=false
RMU_FUNCTION=""
HEAD_STREAM_ARN=""
FAILURE_QUEUE_URL=""
RECOVERY_FUNCTION=""

usage() {
  cat <<'USAGE'
Usage: apply.sh [options]

Required:
  --rmu-function <name>       Lambda function that consumes the head stream
  --head-stream-arn <arn>     DynamoDB Streams ARN of the head table
  --failure-queue-url <url>   SQS queue for OnFailure and recovery notifications
  --recovery-function <name>  Lambda function that runs recoverRmuWorker

Options:
  --apply                     Apply the changes. Without it the script only
                              prints the plan and exits.
  -h, --help                  Show this help.

The script reads event-source-mapping.json next to it. It sets the function
timeout and image command, the queue attributes, both event source mappings,
and the EventBridge schedule that runs the recovery worker when an SQS
notification is lost. It does not change function environment variables.
USAGE
}

require_flag() {
  if [[ -z "$2" ]]; then
    echo "missing required option: $1" >&2
    usage >&2
    exit 2
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --rmu-function)
      require_flag "$1" "${2:-}"
      RMU_FUNCTION="$2"
      shift 2
      ;;
    --head-stream-arn)
      require_flag "$1" "${2:-}"
      HEAD_STREAM_ARN="$2"
      shift 2
      ;;
    --failure-queue-url)
      require_flag "$1" "${2:-}"
      FAILURE_QUEUE_URL="$2"
      shift 2
      ;;
    --recovery-function)
      require_flag "$1" "${2:-}"
      RECOVERY_FUNCTION="$2"
      shift 2
      ;;
    --apply)
      APPLY=true
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

for tool in aws jq; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "required command not found: $tool" >&2
    exit 1
  fi
done

if [[ ! -f "$CONFIG" ]]; then
  echo "config not found: $CONFIG" >&2
  exit 1
fi

value() {
  local result
  result=$(jq -r "$1" "$CONFIG")
  if [[ "$result" == "null" ]]; then
    echo "missing config value: $1" >&2
    exit 1
  fi
  printf '%s\n' "$result"
}

RMU_COMMAND=$(value '.rmuFunction.imageCommand')
RMU_TIMEOUT=$(value '.rmuFunction.timeoutSeconds')
HEAD_BATCH=$(value '.headStreamEventSourceMapping.batchSize')
HEAD_BISECT=$(value '.headStreamEventSourceMapping.bisectBatchOnFunctionError')
HEAD_RESPONSE_TYPES=()
while IFS= read -r response_type; do
  HEAD_RESPONSE_TYPES+=("$response_type")
done < <(jq -r '.headStreamEventSourceMapping.functionResponseTypes[]' "$CONFIG")
if [[ ${#HEAD_RESPONSE_TYPES[@]} -eq 0 ]]; then
  echo "missing config value: .headStreamEventSourceMapping.functionResponseTypes" >&2
  exit 1
fi
HEAD_MAX_RETRY=$(value '.headStreamEventSourceMapping.maximumRetryAttempts')
HEAD_MAX_AGE=$(value '.headStreamEventSourceMapping.maximumRecordAgeInSeconds')
QUEUE_RETENTION=$(value '.failureQueue.messageRetentionPeriodSeconds')
QUEUE_VISIBILITY=$(value '.failureQueue.visibilityTimeoutSeconds')
WORKER_COMMAND=$(value '.recoveryWorkerFunction.imageCommand')
WORKER_TIMEOUT=$(value '.recoveryWorkerFunction.timeoutSeconds')
WORKER_BATCH=$(value '.recoveryWorkerEventSourceMapping.batchSize')
WORKER_RESPONSE_TYPES=()
while IFS= read -r response_type; do
  WORKER_RESPONSE_TYPES+=("$response_type")
done < <(jq -r '.recoveryWorkerEventSourceMapping.functionResponseTypes[]' "$CONFIG")
if [[ ${#WORKER_RESPONSE_TYPES[@]} -eq 0 ]]; then
  echo "missing config value: .recoveryWorkerEventSourceMapping.functionResponseTypes" >&2
  exit 1
fi
WORKER_MAX_RETRY=$(value '.recoveryWorkerEventSourceMapping.maximumRetryAttempts')
SCHEDULE_RULE=$(value '.recoveryWorkerSchedule.ruleName')
SCHEDULE_EXPRESSION=$(value '.recoveryWorkerSchedule.scheduleExpression')
SCHEDULE_STATE=$(value '.recoveryWorkerSchedule.state')
SCHEDULE_MAX_RETRY=$(value '.recoveryWorkerSchedule.maximumRetryAttempts')
SCHEDULE_MAX_AGE=$(value '.recoveryWorkerSchedule.maximumEventAgeSeconds')

if [[ "$HEAD_BISECT" == "true" ]]; then
  BISECT_FLAG="--bisect-batch-on-function-error"
else
  BISECT_FLAG="--no-bisect-batch-on-function-error"
fi

echo "rmu function:        ${RMU_FUNCTION} (timeout ${RMU_TIMEOUT}s, image command ${RMU_COMMAND})"
echo "failure queue:       ${FAILURE_QUEUE_URL} (retention ${QUEUE_RETENTION}s, visibility ${QUEUE_VISIBILITY}s)"
echo "head stream mapping: ${HEAD_STREAM_ARN} (batch ${HEAD_BATCH}, bisect ${HEAD_BISECT}, ${HEAD_RESPONSE_TYPES[*]}, retries ${HEAD_MAX_RETRY}, max age ${HEAD_MAX_AGE}s)"
echo "recovery worker:     ${RECOVERY_FUNCTION} (timeout ${WORKER_TIMEOUT}s, image command ${WORKER_COMMAND})"
echo "worker mapping:      ${FAILURE_QUEUE_URL} (batch ${WORKER_BATCH}, ${WORKER_RESPONSE_TYPES[*]}, retries ${WORKER_MAX_RETRY})"
echo "recovery schedule:   ${SCHEDULE_RULE} (${SCHEDULE_EXPRESSION}, ${SCHEDULE_STATE}, retries ${SCHEDULE_MAX_RETRY}, max age ${SCHEDULE_MAX_AGE}s)"

if [[ "$APPLY" != "true" ]]; then
  echo
  echo "dry-run: no AWS calls were made. Re-run with --apply to apply."
  exit 0
fi

QUEUE_ARN=$(aws sqs get-queue-attributes \
  --queue-url "$FAILURE_QUEUE_URL" \
  --attribute-names QueueArn \
  --query 'Attributes.QueueArn' \
  --output text)

aws sqs set-queue-attributes \
  --queue-url "$FAILURE_QUEUE_URL" \
  --attributes "{\"MessageRetentionPeriod\":\"${QUEUE_RETENTION}\",\"VisibilityTimeout\":\"${QUEUE_VISIBILITY}\"}"

aws lambda update-function-configuration \
  --function-name "$RMU_FUNCTION" \
  --timeout "$RMU_TIMEOUT" \
  --image-config "{\"Command\":[\"${RMU_COMMAND}\"]}"
aws lambda wait function-updated-v2 --function-name "$RMU_FUNCTION"

aws lambda update-function-configuration \
  --function-name "$RECOVERY_FUNCTION" \
  --timeout "$WORKER_TIMEOUT" \
  --image-config "{\"Command\":[\"${WORKER_COMMAND}\"]}"
aws lambda wait function-updated-v2 --function-name "$RECOVERY_FUNCTION"

HEAD_MAPPING_UUID=$(aws lambda list-event-source-mappings \
  --function-name "$RMU_FUNCTION" \
  --event-source-arn "$HEAD_STREAM_ARN" \
  --query 'EventSourceMappings[0].UUID' \
  --output text)

if [[ "$HEAD_MAPPING_UUID" == "None" || -z "$HEAD_MAPPING_UUID" ]]; then
  aws lambda create-event-source-mapping \
    --function-name "$RMU_FUNCTION" \
    --event-source-arn "$HEAD_STREAM_ARN" \
    --batch-size "$HEAD_BATCH" \
    "$BISECT_FLAG" \
    --function-response-types "${HEAD_RESPONSE_TYPES[@]}" \
    --maximum-retry-attempts "$HEAD_MAX_RETRY" \
    --maximum-record-age-in-seconds "$HEAD_MAX_AGE" \
    --destination-config "{\"OnFailure\":{\"Destination\":\"${QUEUE_ARN}\"}}"
else
  aws lambda update-event-source-mapping \
    --uuid "$HEAD_MAPPING_UUID" \
    --batch-size "$HEAD_BATCH" \
    "$BISECT_FLAG" \
    --function-response-types "${HEAD_RESPONSE_TYPES[@]}" \
    --maximum-retry-attempts "$HEAD_MAX_RETRY" \
    --maximum-record-age-in-seconds "$HEAD_MAX_AGE" \
    --destination-config "{\"OnFailure\":{\"Destination\":\"${QUEUE_ARN}\"}}"
fi

WORKER_MAPPING_UUID=$(aws lambda list-event-source-mappings \
  --function-name "$RECOVERY_FUNCTION" \
  --event-source-arn "$QUEUE_ARN" \
  --query 'EventSourceMappings[0].UUID' \
  --output text)

if [[ "$WORKER_MAPPING_UUID" == "None" || -z "$WORKER_MAPPING_UUID" ]]; then
  aws lambda create-event-source-mapping \
    --function-name "$RECOVERY_FUNCTION" \
    --event-source-arn "$QUEUE_ARN" \
    --batch-size "$WORKER_BATCH" \
    --function-response-types "${WORKER_RESPONSE_TYPES[@]}" \
    --maximum-retry-attempts "$WORKER_MAX_RETRY"
else
  aws lambda update-event-source-mapping \
    --uuid "$WORKER_MAPPING_UUID" \
    --batch-size "$WORKER_BATCH" \
    --function-response-types "${WORKER_RESPONSE_TYPES[@]}" \
    --maximum-retry-attempts "$WORKER_MAX_RETRY"
fi

# The schedule is the safety net for a lost SQS notification. The worker
# ignores the event body and replays the pending tasks in MySQL.
RULE_ARN=$(aws events put-rule \
  --name "$SCHEDULE_RULE" \
  --schedule-expression "$SCHEDULE_EXPRESSION" \
  --state "$SCHEDULE_STATE" \
  --query 'RuleArn' \
  --output text)

WORKER_ARN=$(aws lambda get-function-configuration \
  --function-name "$RECOVERY_FUNCTION" \
  --query 'FunctionArn' \
  --output text)

# AWS CLI parses --targets as shorthand. The nested RetryPolicy must use the
# shorthand form; quoted JSON keys fail with "Expected: '=', received: '\"'".
aws events put-targets \
  --rule "$SCHEDULE_RULE" \
  --targets "Id=recovery-worker,Arn=${WORKER_ARN},RetryPolicy={MaximumRetryAttempts=${SCHEDULE_MAX_RETRY},MaximumEventAgeInSeconds=${SCHEDULE_MAX_AGE}}"

if ! PERMISSION_OUTPUT=$(aws lambda add-permission \
  --function-name "$RECOVERY_FUNCTION" \
  --statement-id "$SCHEDULE_RULE" \
  --action lambda:InvokeFunction \
  --principal events.amazonaws.com \
  --source-arn "$RULE_ARN" 2>&1); then
  # add-permission is not idempotent. A statement from an earlier run is fine.
  if [[ "$PERMISSION_OUTPUT" != *ResourceConflictException* ]]; then
    echo "$PERMISSION_OUTPUT" >&2
    exit 1
  fi
fi

echo
echo "applied: ${CONFIG}"
