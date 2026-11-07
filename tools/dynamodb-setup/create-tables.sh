#!/bin/sh

# shellcheck disable=SC2046
cd $(dirname "$0") && pwd

# shellcheck disable=SC2039
if [ $# == 0 ]; then
  echo "Parameters are empty."
  exit 1
fi

while getopts e: OPT; do
  # shellcheck disable=SC2220
  case ${OPT} in
  "e") ENV_NAME="$OPTARG" ;;
  esac
done

export AWS_DEFAULT_REGION=$AWS_REGION
export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-x}
export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-x}
export AWS_PAGER=""

ENDPOINT_URL_OPTION=""
DYNAMODB_ENDPOINT=${DYNAMODB_ENDPOINT:-localhost:31566}

JOURNAL_TABLE_NAME=${JOURNAL_TABLE_NAME:-"journal"}
SNAPSHOT_TABLE_NAME=${SNAPSHOT_TABLE_NAME:-"snapshot"}
HEAD_TABLE_NAME=${HEAD_TABLE_NAME:-"head"}
SNAPSHOT_HISTORY_INDEX_NAME=${SNAPSHOT_HISTORY_INDEX_NAME:-"snapshot-history"}

echo "DYNAMODB_ENDPOINT = ${DYNAMODB_ENDPOINT}"
echo "JOURNAL_TABLE_NAME = ${JOURNAL_TABLE_NAME}"
echo "SNAPSHOT_TABLE_NAME = ${SNAPSHOT_TABLE_NAME}"
echo "HEAD_TABLE_NAME = ${HEAD_TABLE_NAME}"
echo "SNAPSHOT_HISTORY_INDEX_NAME = ${SNAPSHOT_HISTORY_INDEX_NAME}"

if [ "${ENV_NAME}" = "dev" ]; then
  # shellcheck disable=SC2034
  ENDPOINT_URL_OPTION=" --endpoint-url http://${DYNAMODB_ENDPOINT} "
fi

# The journal keeps Streams enabled only until the read model updater moves to
# the head table in #668.
# shellcheck disable=SC2086
aws dynamodb create-table \
  ${ENDPOINT_URL_OPTION} \
  --table-name "${JOURNAL_TABLE_NAME}" \
  --attribute-definitions \
    AttributeName=aid,AttributeType=S \
    AttributeName=seq_nr,AttributeType=N \
  --key-schema \
    AttributeName=aid,KeyType=HASH \
    AttributeName=seq_nr,KeyType=RANGE \
  --provisioned-throughput \
    ReadCapacityUnits=10,WriteCapacityUnits=10 \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE

# shellcheck disable=SC2086
aws dynamodb create-table \
  ${ENDPOINT_URL_OPTION} \
  --table-name "${SNAPSHOT_TABLE_NAME}" \
  --attribute-definitions \
    AttributeName=aid,AttributeType=S \
    AttributeName=skey,AttributeType=N \
    AttributeName=active_history_seq_nr,AttributeType=N \
  --key-schema \
    AttributeName=aid,KeyType=HASH \
    AttributeName=skey,KeyType=RANGE \
  --provisioned-throughput \
    ReadCapacityUnits=10,WriteCapacityUnits=10 \
  --global-secondary-indexes \
  "[
    {
      \"IndexName\": \"${SNAPSHOT_HISTORY_INDEX_NAME}\",
      \"KeySchema\": [{\"AttributeName\":\"aid\",\"KeyType\":\"HASH\"},
                      {\"AttributeName\":\"active_history_seq_nr\",\"KeyType\":\"RANGE\"}],
      \"Projection\":{
        \"ProjectionType\":\"KEYS_ONLY\"
      },
      \"ProvisionedThroughput\": {
        \"ReadCapacityUnits\": 10,
        \"WriteCapacityUnits\": 10
      }
    }
  ]"

# shellcheck disable=SC2086
aws dynamodb create-table \
  ${ENDPOINT_URL_OPTION} \
  --table-name "${HEAD_TABLE_NAME}" \
  --attribute-definitions \
    AttributeName=aid,AttributeType=S \
  --key-schema \
    AttributeName=aid,KeyType=HASH \
  --provisioned-throughput \
    ReadCapacityUnits=10,WriteCapacityUnits=10 \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE
