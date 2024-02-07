#!/usr/bin/env bash

curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getMessage(messageId: \"${MESSAGE_ID}\", userAccountId: \"${USER_ACCOUNT_ID}\") { id, groupChatId, text, createdAt } }" }
EOS
