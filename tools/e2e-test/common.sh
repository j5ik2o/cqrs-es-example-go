#!/usr/bin/env bash

# Shared GraphQL helpers for the docker-compose e2e scripts.

ADMIN_ID=${ADMIN_ID:-UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z}
USER_ACCOUNT_ID=${USER_ACCOUNT_ID:-UserAccount-01H7C6DWMK1BKS1JYH1XZE529M}
WRITE_API_SERVER_BASE_URL=${WRITE_API_SERVER_BASE_URL:-http://localhost:28080}
READ_API_SERVER_BASE_URL=${READ_API_SERVER_BASE_URL:-http://localhost:28082}
POLL_ATTEMPTS=${POLL_ATTEMPTS:-90}
POLL_INTERVAL_SECONDS=${POLL_INTERVAL_SECONDS:-2}

MUTATION_RESULT=""

post_graphql() {
	local base_url="$1"
	local body="$2"
	curl -sS --max-time 30 -X POST -H "Content-Type: application/json" "${base_url}/query" -d "${body}"
}

# wait_for_api <label> <base_url>. Polls until the GraphQL endpoint answers an
# HTTP request. Any HTTP status counts: the e2e scripts want the server to be
# listening before the first mutation, and the read model itself is polled
# separately by wait_for_read.
wait_for_api() {
	local label="$1"
	local base_url="$2"
	local attempt=0
	while [ "${attempt}" -lt "${POLL_ATTEMPTS}" ]; do
		if curl -sS --max-time 5 -o /dev/null -X POST -H "Content-Type: application/json" \
			"${base_url}/query" -d '{"query":"{ __typename }"}' 2>/dev/null; then
			echo "${label} is ready: ${base_url}"
			return 0
		fi
		attempt=$((attempt + 1))
		sleep "${POLL_INTERVAL_SECONDS}"
	done
	echo "Error: timed out after ${POLL_ATTEMPTS} attempts waiting for ${label}: ${base_url}" >&2
	return 1
}

# run_mutation <label> <body>. Stores the raw response in MUTATION_RESULT.
run_mutation() {
	local label="$1"
	local body="$2"
	local result
	echo -e "\n${label}:"
	result=$(post_graphql "${WRITE_API_SERVER_BASE_URL}" "${body}")
	if echo "${result}" | jq -e '.errors' > /dev/null 2>&1; then
		echo "Error: ${result}"
		exit 1
	fi
	if ! echo "${result}" | jq -e '.data != null' > /dev/null 2>&1; then
		echo "Error: mutation returned no data: ${result}"
		exit 1
	fi
	echo "Result: ${result}"
	MUTATION_RESULT="${result}"
}

# wait_for_read <label> <body> <jq-filter>. Polls the read model until the
# projection catches up, so a fixed sleep is not the only thing separating a
# write from its read.
wait_for_read() {
	local label="$1"
	local body="$2"
	local filter="$3"
	local result=""
	local attempt=0
	echo -e "\n${label}:"
	while [ "${attempt}" -lt "${POLL_ATTEMPTS}" ]; do
		result=$(post_graphql "${READ_API_SERVER_BASE_URL}" "${body}")
		if echo "${result}" | jq -e '.errors' > /dev/null 2>&1; then
			# A 404 means the RMU has not projected the event yet; keep polling.
			if echo "${result}" | jq -e '.errors | all(.extensions.code == "404")' > /dev/null 2>&1; then
				attempt=$((attempt + 1))
				sleep "${POLL_INTERVAL_SECONDS}"
				continue
			fi
			echo "Error: ${result}"
			exit 1
		fi
		if echo "${result}" | jq -e "${filter}" > /dev/null 2>&1; then
			echo "Result: ${result}"
			return 0
		fi
		attempt=$((attempt + 1))
		sleep "${POLL_INTERVAL_SECONDS}"
	done
	echo "Error: timed out after ${POLL_ATTEMPTS} attempts waiting for ${label}: ${result}"
	exit 1
}
