# cqrs-es-example-go

## Overview

This is an example of CQRS/Event Sourcing implemented in Go.

This project uses [j5ik2o/event-store-adapter-go](https://github.com/j5ik2o/event-store-adapter-go) for Event Sourcing.

[日本語](./README.ja.md)

## Feature

- [x] Write API Server(REST)
- [x] Read API Server(GraphQL)
- [x] Read Model Updater on Local
- [x] Docker Compose Support
- [ ] Read Model Updater on AWS Lambda
- [ ] Deployment to AWS

## Overview

### Component Composition

- Write API Server
  - Write-only Web API
- Read Model Updater
  - Lambda to build read models based on journals
- Read API Server
  - implemented by GraphQL (Query, Subscription)

### System Architecture Diagram

![](docs/images/system-layout.png)

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

#### Testing

```shell
$ make verify-group-chat
ADMIN_ID="UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z" \
        WRITE_API_SERVER_BASE_URL=http://localhost:28080 \
        READ_API_SERVER_BASE_URL=http://localhost:28082 \
        ./tools/scripts/verify-group-chat.sh
{"group_chat_id":"GroupChat-01HPG4EV94HMPT08GZS0ZWW0VJ"}
GroupChat:
{
  "data": {
    "getGroupChat": {
      "id": "GroupChat-01HPG4EV94HMPT08GZS0ZWW0VJ",
      "name": "group-chat-example-1",
      "ownerId": "UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z",
      "createdAt": "2024-02-13 02:24:12 +0000 UTC",
      "updatedAt": "2024-02-13 02:24:12 +0000 UTC"
    }
  }
}
...
```

## Links

- [Rust Version](https://github.com/j5ik2o/cqrs-es-example-rs)
- [TypeScript Version](https://github.com/j5ik2o/cqrs-es-example-js)
- [Common Documents](https://github.com/j5ik2o/cqrs-es-example-docs)
