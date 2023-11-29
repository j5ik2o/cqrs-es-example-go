#!/usr/bin/env bash

curl -s -X 'POST' \
    "${WRITE_API_SERVER_BASE_URL}/group-chats/add-member" \
    -H 'accept: application/json' \
    -H 'Content-Type: application/json' \
    -d "{ \"executor_id\": \"01H42K4ABWQ5V2XQEP3A48VE0Z\", \"group_chat_id\": \"${GROUP_CHAT_ID}\", \"account_id\": \"${ACCOUNT_ID}\", \"role\": \"${ROLE}\" }"
