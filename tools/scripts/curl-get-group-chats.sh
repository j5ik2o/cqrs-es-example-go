#!/usr/bin/env bash

curl -s -X POST -H "Content-Type: application/json" \
	${READ_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{ "query": "{ getGroupChats(userAccountId: \"${ADMIN_ID}\") { id, name, ownerId, createdAt } }" }
EOS
