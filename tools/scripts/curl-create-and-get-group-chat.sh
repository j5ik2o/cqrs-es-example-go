#!/usr/bin/env bash

export GROUP_CHAT_ID=$(./tools/scripts/curl-create-group-chat.sh | jq -r .group_chat_id)
echo "create-group-chat: GROUP_CHAT_ID=${GROUP_CHAT_ID}"

# トライの最大回数
MAX_RETRIES=10
# トライごとの待ち時間
SLEEP_TIME=1

for i in $(seq 1 $MAX_RETRIES); do
    ACTUAL_ID=$(./tools/scripts/curl-get-group-chat.sh | jq -r .data.getGroupChat.id)
    echo "get-group-chat (attempt ${i}): ACTUAL_GROUP_CHAT_ID=${ACTUAL_ID}"

    if [ "${GROUP_CHAT_ID}" = "${ACTUAL_ID}" ]; then
        echo "OK"
        exit 0
    fi

    # 指定された時間だけ待つ
    sleep $SLEEP_TIME
done

echo "NG"
exit 1
