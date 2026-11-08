FROM bash:devel-alpine3.19

RUN apk add --no-cache curl jq aws-cli

COPY ./common.sh ./verify-group-chat.sh ./verify-after-restart.sh ./verify-deploy-assets.sh ./

RUN chmod +x ./verify-group-chat.sh ./verify-after-restart.sh ./verify-deploy-assets.sh

ENTRYPOINT ["./verify-group-chat.sh"]
