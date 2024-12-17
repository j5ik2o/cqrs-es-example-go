# syntax=docker/dockerfile:1
FROM golang:1.23 AS build
ARG TARGETARCH
COPY . /app
WORKDIR /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o /cqrs-es-example-go

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /cqrs-es-example-go /cqrs-es-example-go

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/cqrs-es-example-go"]
