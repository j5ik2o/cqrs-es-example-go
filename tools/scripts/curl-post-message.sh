#!/usr/bin/env bash

curl -s -X POST -H "Content-Type: application/json" \
	${WRITE_API_SERVER_BASE_URL}/query \
	-d @- <<EOS
{
  "query": "mutation PostMessage(\$input: PostMessageInput!) { postMessage(input: \$input) { groupChatId, messageId } }",
  "variables": {
    "input": {
      "groupChatId": "${GROUP_CHAT_ID}",
      "content": "Text1",
      "userAccountId": "${USER_ACCOUNT_ID}",
      "executorId": "${USER_ACCOUNT_ID}"
    }
  }
}
EOS
