#!/usr/bin/env bash

WRITE_API_SERVER_BASE_URL=http://localhost:18080
READ_API_SERVER_BASE_URL=http://localhost:18082
ADMIN_ID=01H42K4ABWQ5V2XQEP3A48VE0Z
USER_ACCOUNT_ID=01H7C6DWMK1BKS1JYH1XZE529M

# グループチャット作成
GROUP_CHAT_ID=$(curl -s -X 'POST' \
  "${WRITE_API_SERVER_BASE_URL}/group-chats/create" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{ \"name\": \"group-chat-example-1\", \"executor_id\": \"${ADMIN_ID}\" }" | jq -r .group_chat_id)

# メンバー追加
curl -v -X 'POST' \
  "${WRITE_API_SERVER_BASE_URL}/group-chats/add-member" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{ \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"role\": \"member\", \"user_account_id\": \"${USER_ACCOUNT_ID}\", \"executor_id\": \"${ADMIN_ID}\" }"

# メッセージ投稿1
MESSAGE_ID=$(curl -v -X 'POST' \
  "${WRITE_API_SERVER_BASE_URL}/group-chats/post-message" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{ \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"message\": \"Text1\", \"user_account_id\": \"${USER_ACCOUNT_ID}\", \"executor_id\": \"${USER_ACCOUNT_ID}\"  }" \
  | jq -r .message_id)

echo "MESSAGE_ID: ${MESSAGE_ID}"
echo "USER_ACCOUNT_ID: ${USER_ACCOUNT_ID}"

curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getMessage(messageId: \"${MESSAGE_ID}\", userAccountId: \"UserAccount-${USER_ACCOUNT_ID}\") { id, groupChatId, text, createdAt } }" }
EOS
