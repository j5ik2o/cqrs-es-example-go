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
curl -s -X 'POST' \
  "${WRITE_API_SERVER_BASE_URL}/group-chats/add-member" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{ \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"role\": \"member\", \"user_account_id\": \"${USER_ACCOUNT_ID}\", \"executor_id\": \"${ADMIN_ID}\" }"

# メッセージ投稿1
MESSAGE_ID=$(curl -s -X 'POST' \
  "${WRITE_API_SERVER_BASE_URL}/group-chats/post-message" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{ \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"message\": \"Text1\", \"user_account_id\": \"${USER_ACCOUNT_ID}\", \"executor_id\": \"${USER_ACCOUNT_ID}\"  }" \
  | jq -r .message_id)

# グループチャット取得
group_chat=$(curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getGroupChat(groupChatId: \"${GROUP_CHAT_ID}\", userAccountId: \"UserAccount-${ADMIN_ID}\") { id, name, ownerId, createdAt, updatedAt } }" }
EOS)

echo -e "\nGroupChat:"
echo $group_chat | jq .

# グループチャットリスト取得
group_list_chat=$(curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getGroupChats(userAccountId: \"UserAccount-${ADMIN_ID}\") { id, name, ownerId, createdAt, updatedAt } }" }
EOS)

echo -e "\nGroupChats:"
echo $group_list_chat | jq .

# メンバー取得
member=$(curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS1
{ "query": "{ getMember(groupChatId: \"${GROUP_CHAT_ID}\", userAccountId: \"UserAccount-${USER_ACCOUNT_ID}\") { id, groupChatId, userAccountId, role, createdAt, updatedAt } }" }
EOS1)

echo -e "\nMember:"
echo $member | jq .

# メンバーリスト取得
member_list=$(curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS3
{ "query": "{ getMembers(groupChatId: \"${GROUP_CHAT_ID}\", userAccountId: \"UserAccount-${USER_ACCOUNT_ID}\") { id, groupChatId, userAccountId, role, createdAt, updatedAt } }" }
EOS3)

echo -e "\nMembers:"
echo $member_list | jq .

# メッセージ取得
message=$(curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS2
{ "query": "{ getMessage(messageId: \"${MESSAGE_ID}\", userAccountId: \"UserAccount-${USER_ACCOUNT_ID}\") { id, groupChatId, text, createdAt, updatedAt } }" }
EOS2)

echo -e "\nMessage:"
echo $message | jq .

# メッセージリスト取得
message_list=$(curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS3
{ "query": "{ getMessages(groupChatId: \"${GROUP_CHAT_ID}\", userAccountId: \"UserAccount-${USER_ACCOUNT_ID}\") { id, groupChatId, text, createdAt, updatedAt } }" }
EOS3)

echo -e "\nMessages:"
echo $message_list | jq .
