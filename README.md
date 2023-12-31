# cqrs-es-example-go

## Overview

This is an example of CQRS/Event Sourcing implemented in Go.

## Usage

### Local Environment

#### Starting Services

```shell
$ make docker-compose-up
./tools/scripts/docker-compose-up.sh
ARCH=arm64
 Container read-api-server-1  Stopping
...
 Container docker-compose-migration-1  Started
 Container read-api-server-1  Starting
 Container read-api-server-1  Started
```

#### Creating a Group Chat

```shell
$ make create-group-chat
WRITE_API_SERVER_BASE_URL=http://localhost:18080 ./tools/scripts/curl-post-create-group-chat.sh
{"group_chat_id":"GroupChat-01HGDR3R6KFHRHTF0K54GMW5XF"}%
```

#### Getting a Group Chat

```shell
$ GROUP_CHAT_ID=GroupChat-01HGDR3R6KFHRHTF0K54GMW5XF make get-group-chat
{"data":{"getGroupChat":{"id":"GroupChat-01HGDR3R6KFHRHTF0K54GMW5XF"}}}%
```
