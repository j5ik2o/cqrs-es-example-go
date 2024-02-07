#!/usr/bin/env bash

curl -s -X 'POST' \
  "${WRITE_API_SERVER_BASE_URL}/group-chats/post-message" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{ \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"message\": \"Text1\", \"user_account_id\": \"${USER_ACCOUNT_ID}\", \"executor_id\": \"${USER_ACCOUNT_ID}\" }"
