#!/usr/bin/env bash
#
# Apply the RMU recovery alarms and their CloudWatch Logs metric filters from
# cloudwatch-alarms.json. The script is a dry run unless --apply is passed.
# Read-only lookups are skipped in a dry run, so a plan needs no credentials.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CONFIG="${SCRIPT_DIR}/cloudwatch-alarms.json"

APPLY=false
RMU_FUNCTION=""
RECOVERY_FUNCTION=""
FAILURE_QUEUE_URL=""
ALARM_ACTIONS=""

usage() {
  cat <<'USAGE'
Usage: apply-alarms.sh [options]

Required:
  --rmu-function <name>       Lambda function that consumes the head stream
  --recovery-function <name>  Lambda function that runs recoverRmuWorker
  --failure-queue-url <url>   SQS queue for OnFailure and recovery notifications

Options:
  --alarm-actions <arn>       SNS topic ARN to notify when an alarm fires.
                              Without it the alarms only change state in the
                              console.
  --apply                     Apply the changes. Without it the script only
                              prints the plan and exits.
  -h, --help                  Show this help.

The script reads cloudwatch-alarms.json next to it. It creates the recovery
worker's log group when it does not exist, then applies the metric filters and
alarms. It does not change the functions, the queue, or the event source
mappings; run apply.sh for those.
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
    --recovery-function)
      require_flag "$1" "${2:-}"
      RECOVERY_FUNCTION="$2"
      shift 2
      ;;
    --failure-queue-url)
      require_flag "$1" "${2:-}"
      FAILURE_QUEUE_URL="$2"
      shift 2
      ;;
    --alarm-actions)
      require_flag "$1" "${2:-}"
      ALARM_ACTIONS="$2"
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

NAMESPACE=$(jq -r '.namespace' "$CONFIG")
LOG_GROUP="/aws/lambda/${RECOVERY_FUNCTION}"
QUEUE_NAME="${FAILURE_QUEUE_URL##*/}"

echo "log group:  ${LOG_GROUP}"
echo "namespace:  ${NAMESPACE}"
jq -r '.metricFilters[] | "filter:     \(.name) -> \(.metricName) from \(.metricValue)"' "$CONFIG"
jq -r '.alarms[] | "alarm:      \(.name) \(.namespace)/\(.metricName) \(.comparisonOperator) \(.threshold)"' "$CONFIG"
if [[ -n "$ALARM_ACTIONS" ]]; then
  echo "actions:    ${ALARM_ACTIONS}"
fi

if [[ "$APPLY" != "true" ]]; then
  echo
  echo "dry-run: no AWS calls were made. Re-run with --apply to apply."
  exit 0
fi

# Lambda creates the log group on the first invocation, but the metric filters
# must work even before that. create-log-group is not idempotent, so an
# existing group is the expected failure.
if ! CREATE_OUTPUT=$(aws logs create-log-group --log-group-name "$LOG_GROUP" 2>&1); then
  if [[ "$CREATE_OUTPUT" != *ResourceAlreadyExistsException* ]]; then
    echo "$CREATE_OUTPUT" >&2
    exit 1
  fi
fi

FILTER_COUNT=$(jq '.metricFilters | length' "$CONFIG")
for ((i = 0; i < FILTER_COUNT; i++)); do
  FILTER_NAME=$(jq -r ".metricFilters[$i].name" "$CONFIG")
  PATTERN=$(jq -r ".metricFilters[$i].pattern" "$CONFIG")
  METRIC_NAME=$(jq -r ".metricFilters[$i].metricName" "$CONFIG")
  METRIC_VALUE=$(jq -r ".metricFilters[$i].metricValue" "$CONFIG")
  DEFAULT_VALUE=$(jq -r ".metricFilters[$i].defaultValue" "$CONFIG")
  aws logs put-metric-filter \
    --log-group-name "$LOG_GROUP" \
    --filter-name "$FILTER_NAME" \
    --filter-pattern "$PATTERN" \
    --metric-transformations "metricName=${METRIC_NAME},metricNamespace=${NAMESPACE},metricValue=${METRIC_VALUE},defaultValue=${DEFAULT_VALUE}"
done

ALARM_COUNT=$(jq '.alarms | length' "$CONFIG")
for ((i = 0; i < ALARM_COUNT; i++)); do
  ALARM=$(jq -c ".alarms[$i]" "$CONFIG")
  ALARM_NAME=$(jq -r '.name' <<<"$ALARM")
  DESCRIPTION=$(jq -r '.description' <<<"$ALARM")
  ALARM_NAMESPACE=$(jq -r '.namespace' <<<"$ALARM")
  METRIC_NAME=$(jq -r '.metricName' <<<"$ALARM")
  STATISTIC=$(jq -r '.statistic' <<<"$ALARM")
  PERIOD=$(jq -r '.periodSeconds' <<<"$ALARM")
  EVALUATION_PERIODS=$(jq -r '.evaluationPeriods' <<<"$ALARM")
  DATAPOINTS=$(jq -r '.datapointsToAlarm' <<<"$ALARM")
  THRESHOLD=$(jq -r '.threshold' <<<"$ALARM")
  COMPARISON=$(jq -r '.comparisonOperator' <<<"$ALARM")
  MISSING=$(jq -r '.treatMissingData' <<<"$ALARM")
  DIMENSIONS=$(jq -c \
    --arg rmuFunction "$RMU_FUNCTION" \
    --arg recoveryFunction "$RECOVERY_FUNCTION" \
    --arg failureQueueName "$QUEUE_NAME" '
      [.dimensions[]? | {
        Name: .name,
        Value: (
          if .valueFrom == "rmuFunction" then $rmuFunction
          elif .valueFrom == "recoveryFunction" then $recoveryFunction
          elif .valueFrom == "failureQueueName" then $failureQueueName
          else error("unknown valueFrom: \(.valueFrom)")
          end
        )
      }]' <<<"$ALARM")

  ALARM_ARGS=(
    --alarm-name "$ALARM_NAME"
    --alarm-description "$DESCRIPTION"
    --namespace "$ALARM_NAMESPACE"
    --metric-name "$METRIC_NAME"
    --statistic "$STATISTIC"
    --period "$PERIOD"
    --evaluation-periods "$EVALUATION_PERIODS"
    --threshold "$THRESHOLD"
    --comparison-operator "$COMPARISON"
    --treat-missing-data "$MISSING"
  )
  if [[ "$DATAPOINTS" != "null" && "$DATAPOINTS" != "0" ]]; then
    ALARM_ARGS+=(--datapoints-to-alarm "$DATAPOINTS")
  fi
  if [[ "$DIMENSIONS" != "[]" ]]; then
    ALARM_ARGS+=(--dimensions "$DIMENSIONS")
  fi
  if [[ -n "$ALARM_ACTIONS" ]]; then
    ALARM_ARGS+=(--alarm-actions "$ALARM_ACTIONS")
  fi
  aws cloudwatch put-metric-alarm "${ALARM_ARGS[@]}"
done

echo
echo "applied: ${CONFIG}"
