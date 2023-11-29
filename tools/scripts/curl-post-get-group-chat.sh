#!/usr/bin/env bash

curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getGroupChat(groupChatId: \"${GROUP_CHAT_ID}\", accountId: \"user-account-01H42K4ABWQ5V2XQEP3A48VE0Z\") { id } }" }
EOS
