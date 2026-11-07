#!/usr/bin/env bash

set -ue

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
	curl -s -X POST -H "Content-Type: application/json" "${base_url}/query" -d "${body}"
}

# run_mutation <label> <body>. Stores the raw response in MUTATION_RESULT.
run_mutation() {
	local label="$1"
	local body="$2"
	echo -e "\n${label}:"
	local result
	result=$(post_graphql "${WRITE_API_SERVER_BASE_URL}" "${body}")
	if echo "${result}" | jq -e .errors > /dev/null; then
		echo "Error: ${result}"
		exit 1
	fi
	echo "Result: ${result}"
	MUTATION_RESULT="${result}"
}

# wait_for_read <label> <body> <jq-filter>. Polls the read model so a fixed sleep
# is not the only thing separating a write from its projection.
wait_for_read() {
	local label="$1"
	local body="$2"
	local filter="$3"
	echo -e "\n${label}:"
	local result=""
	local attempt=0
	while [ "${attempt}" -lt "${POLL_ATTEMPTS}" ]; do
		result=$(post_graphql "${READ_API_SERVER_BASE_URL}" "${body}")
		if echo "${result}" | jq -e .errors > /dev/null; then
			# A 404 means the RMU has not projected the event yet; keep polling.
			if echo "${result}" | jq -e '.errors | all(.extensions.code == "404")' > /dev/null; then
				attempt=$((attempt + 1))
				sleep "${POLL_INTERVAL_SECONDS}"
				continue
			fi
			echo "Error: ${result}"
			exit 1
		fi
		if echo "${result}" | jq -e "${filter}" > /dev/null; then
			echo "Result: ${result}"
			return 0
		fi
		attempt=$((attempt + 1))
		sleep "${POLL_INTERVAL_SECONDS}"
	done
	echo "Error: timed out after ${POLL_ATTEMPTS} attempts waiting for ${label}: ${result}"
	exit 1
}

# グループチャット作成
run_mutation "Create GroupChat(${ADMIN_ID})" "$(cat <<EOS
{
  "query": "mutation CreateGroupChat(\$input: CreateGroupChatInput!) { createGroupChat(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "name": "group-chat-example",
      "executorId": "${ADMIN_ID}"
    }
  }
}
EOS
)"
GROUP_CHAT_ID=$(echo "${MUTATION_RESULT}" | jq -r .data.createGroupChat.groupChatId)

# メンバー追加
run_mutation "Add Member(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID}, ${ADMIN_ID})" "$(cat <<EOS
{
  "query": "mutation AddMember(\$input: AddMemberInput!) { addMember(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "userAccountId": "${USER_ACCOUNT_ID}",
      "role": "MEMBER",
      "executorId": "${ADMIN_ID}"
    }
  }
}
EOS
)"

# メッセージ投稿
run_mutation "Post Message(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" "$(cat <<EOS
{
  "query": "mutation PostMessage(\$input: PostMessageInput!) { postMessage(input: \$input) { groupChatId, messageId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "content": "Text1",
      "executorId": "${USER_ACCOUNT_ID}"
    }
  }
}
EOS
)"
MESSAGE_ID=$(echo "${MUTATION_RESULT}" | jq -r .data.postMessage.messageId)

# グループチャット取得
wait_for_read "Get GroupChat(${GROUP_CHAT_ID}, ${ADMIN_ID})" \
	"{ \"query\": \"{ getGroupChat(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${ADMIN_ID}\\\") { id, name, ownerId, createdAt, updatedAt } }\" }" \
	".data.getGroupChat.id == \"${GROUP_CHAT_ID}\""

# グループチャットリスト取得
wait_for_read "Get GroupChats(${ADMIN_ID})" \
	"{ \"query\": \"{ getGroupChats(userAccountId: \\\"${ADMIN_ID}\\\") { id, name, ownerId, createdAt, updatedAt } }\" }" \
	".data.getGroupChats | map(.id) | index(\"${GROUP_CHAT_ID}\") != null"

# メンバー取得
wait_for_read "Get Member(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMember(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, groupChatId, userAccountId, role, createdAt, updatedAt } }\" }" \
	".data.getMember.userAccountId == \"${USER_ACCOUNT_ID}\""

# メンバーリスト取得
wait_for_read "Get Members(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMembers(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, groupChatId, userAccountId, role, createdAt, updatedAt } }\" }" \
	".data.getMembers | length >= 2"

# メッセージ取得
wait_for_read "Get Message(${MESSAGE_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMessage(messageId: \\\"${MESSAGE_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, groupChatId, text, createdAt, updatedAt } }\" }" \
	".data.getMessage.text == \"Text1\""

# メッセージリスト取得
wait_for_read "Get Messages(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMessages(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, groupChatId, text, createdAt, updatedAt } }\" }" \
	".data.getMessages | map(.id) | index(\"${MESSAGE_ID}\") != null"

# グループチャット名の変更
run_mutation "Rename GroupChat(${GROUP_CHAT_ID}, ${ADMIN_ID})" "$(cat <<EOS
{
  "query": "mutation RenameGroupChat(\$input: RenameGroupChatInput!) { renameGroupChat(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "name": "group-chat-example-2",
      "executorId": "${ADMIN_ID}"
    }
  }
}
EOS
)"

wait_for_read "Get GroupChat(${GROUP_CHAT_ID}, ${ADMIN_ID})" \
	"{ \"query\": \"{ getGroupChat(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${ADMIN_ID}\\\") { id, name, ownerId, createdAt, updatedAt } }\" }" \
	".data.getGroupChat.name == \"group-chat-example-2\""

# メッセージ編集
run_mutation "Edit Message(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" "$(cat <<EOS
{
  "query": "mutation EditMessage(\$input: EditMessageInput!) { editMessage(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "messageId": "${MESSAGE_ID}",
      "content": "Text2",
      "executorId": "${USER_ACCOUNT_ID}"
    }
  }
}
EOS
)"

wait_for_read "Get Message(${MESSAGE_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMessage(messageId: \\\"${MESSAGE_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, groupChatId, text, createdAt, updatedAt } }\" }" \
	".data.getMessage.text == \"Text2\""

# メッセージの削除
run_mutation "Delete Message(${GROUP_CHAT_ID}, ${MESSAGE_ID}, ${USER_ACCOUNT_ID})" "$(cat <<EOS
{
  "query": "mutation DeleteMessage(\$input: DeleteMessageInput!) { deleteMessage(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "messageId": "${MESSAGE_ID}",
      "executorId": "${USER_ACCOUNT_ID}"
    }
  }
}
EOS
)"

wait_for_read "Get Messages(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMessages(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, groupChatId, text, createdAt, updatedAt } }\" }" \
	".data.getMessages | map(.id) | index(\"${MESSAGE_ID}\") == null"

# メンバーの削除
run_mutation "Remove Member(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID}, ${ADMIN_ID})" "$(cat <<EOS
{
  "query": "mutation RemoveMember(\$input: RemoveMemberInput!) { removeMember(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "userAccountId": "${USER_ACCOUNT_ID}",
      "executorId": "${ADMIN_ID}"
    }
  }
}
EOS
)"

wait_for_read "Get Members(${GROUP_CHAT_ID}, ${ADMIN_ID})" \
	"{ \"query\": \"{ getMembers(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${ADMIN_ID}\\\") { id, groupChatId, userAccountId, role, createdAt, updatedAt } }\" }" \
	".data.getMembers | map(.userAccountId) | index(\"${USER_ACCOUNT_ID}\") == null"

# ルームの削除
run_mutation "Delete GroupChat(${GROUP_CHAT_ID}, ${ADMIN_ID})" "$(cat <<EOS
{
  "query": "mutation DeleteGroupChat(\$input: DeleteGroupChatInput!) { deleteGroupChat(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "executorId": "${ADMIN_ID}"
    }
  }
}
EOS
)"

wait_for_read "Get GroupChats(${ADMIN_ID})" \
	"{ \"query\": \"{ getGroupChats(userAccountId: \\\"${ADMIN_ID}\\\") { id, name, ownerId, createdAt, updatedAt } }\" }" \
	".data.getGroupChats | map(.id) | index(\"${GROUP_CHAT_ID}\") == null"
