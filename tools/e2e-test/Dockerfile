FROM bash:devel-alpine3.19

RUN apk add --no-cache curl jq

ADD ./common.sh ./verify-group-chat.sh ./verify-after-restart.sh ./

RUN chmod +x ./verify-group-chat.sh ./verify-after-restart.sh

ENTRYPOINT ["./verify-group-chat.sh"]
