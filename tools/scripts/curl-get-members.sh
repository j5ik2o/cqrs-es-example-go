#!/usr/bin/env bash

curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getMembers(groupChatId: \"${GROUP_CHAT_ID}\", userAccountId: \"UserAccount-${USER_ACCOUNT_ID}\") { id, groupChatId, userAccountId, role, createdAt } }" }
EOS
