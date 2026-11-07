#!/usr/bin/env bash

set -u

source "$(dirname "$0")/common.sh"

echo "Waiting for the write and read APIs to accept requests"
wait_for_api "write api" "${WRITE_API_SERVER_BASE_URL}" || exit 1
wait_for_api "read api" "${READ_API_SERVER_BASE_URL}" || exit 1

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
if [ -z "${GROUP_CHAT_ID}" ] || [ "${GROUP_CHAT_ID}" = "null" ]; then
	echo "Error: createGroupChat returned no groupChatId: ${MUTATION_RESULT}"
	exit 1
fi

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
if [ -z "${MESSAGE_ID}" ] || [ "${MESSAGE_ID}" = "null" ]; then
	echo "Error: postMessage returned no messageId: ${MUTATION_RESULT}"
	exit 1
fi

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
