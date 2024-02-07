#!/usr/bin/env bash

curl -s -X 'POST' \
    "${WRITE_API_SERVER_BASE_URL}/group-chats/create" \
    -H 'accept: application/json' \
    -H 'Content-Type: application/json' \
    -d "{\"name\": \"group-chat-example\", \"executor_id\": \"${ADMIN_ID}\"}"
