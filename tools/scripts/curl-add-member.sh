#!/usr/bin/env bash

curl -s -X 'POST' \
    "${WRITE_API_SERVER_BASE_URL}/group-chats/add-member" \
    -H 'accept: application/json' \
    -H 'Content-Type: application/json' \
    -d "{ \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"user_account_id\": \"${USER_ACCOUNT_ID}\", \"role\": \"${ROLE}\", \"executor_id\": \"${ADMIN_ID}\" }"
