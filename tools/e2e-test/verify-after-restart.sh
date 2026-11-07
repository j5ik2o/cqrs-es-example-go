#!/usr/bin/env bash

# Runs after the read-model-updater has been restarted. The fresh write is a
# barrier: head stream records are ordered, so once this event is projected the
# updater has replayed every record written before the restart.

set -u

source "$(dirname "$0")/common.sh"

echo "Waiting for the write and read APIs to accept requests"
wait_for_api "write api" "${WRITE_API_SERVER_BASE_URL}" || exit 1
wait_for_api "read api" "${READ_API_SERVER_BASE_URL}" || exit 1

# 再起動後の新しいグループチャット
run_mutation "Create GroupChat after RMU restart(${ADMIN_ID})" "$(cat <<EOS
{
  "query": "mutation CreateGroupChat(\$input: CreateGroupChatInput!) { createGroupChat(input: \$input) { groupChatId } }",
  "variables": {
    "input": {
      "name": "group-chat-after-restart",
      "executorId": "${ADMIN_ID}"
    }
  }
}
EOS
)"
GROUP_CHAT_ID=$(echo "${MUTATION_RESULT}" | jq -r .data.createGroupChat.groupChatId)
if [ -z "${GROUP_CHAT_ID}" ] || [ "${GROUP_CHAT_ID}" = "null" ]; then
	echo "Error: createGroupChat returned no groupChatId: ${MUTATION_RESULT}"
	exit 1
fi

run_mutation "Add Member after RMU restart(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID}, ${ADMIN_ID})" "$(cat <<EOS
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

run_mutation "Post Message after RMU restart(${GROUP_CHAT_ID}, ${USER_ACCOUNT_ID})" "$(cat <<EOS
{
  "query": "mutation PostMessage(\$input: PostMessageInput!) { postMessage(input: \$input) { groupChatId, messageId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "content": "Text-after-restart",
      "executorId": "${USER_ACCOUNT_ID}"
    }
  }
}
EOS
)"
MESSAGE_ID=$(echo "${MUTATION_RESULT}" | jq -r .data.postMessage.messageId)
if [ -z "${MESSAGE_ID}" ] || [ "${MESSAGE_ID}" = "null" ]; then
	echo "Error: postMessage returned no messageId: ${MUTATION_RESULT}"
	exit 1
fi

wait_for_read "Get GroupChat after RMU restart(${GROUP_CHAT_ID}, ${ADMIN_ID})" \
	"{ \"query\": \"{ getGroupChat(groupChatId: \\\"${GROUP_CHAT_ID}\\\", userAccountId: \\\"${ADMIN_ID}\\\") { id, name } }\" }" \
	".data.getGroupChat.name == \"group-chat-after-restart\""

wait_for_read "Get Message after RMU restart(${MESSAGE_ID}, ${USER_ACCOUNT_ID})" \
	"{ \"query\": \"{ getMessage(messageId: \\\"${MESSAGE_ID}\\\", userAccountId: \\\"${USER_ACCOUNT_ID}\\\") { id, text } }\" }" \
	".data.getMessage.text == \"Text-after-restart\""

echo "AFTER_RESTART_GROUP_CHAT_ID=${GROUP_CHAT_ID}"
echo "AFTER_RESTART_MESSAGE_ID=${MESSAGE_ID}"
