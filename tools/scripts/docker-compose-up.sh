#!/usr/bin/env bash

set -eu

# shellcheck disable=SC2046
cd $(dirname "$0") || exit

export ARCH=$(uname -m)
echo "ARCH=${ARCH}"

if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
fi

if [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

F_OPTION="-f ../docker-compose/docker-compose-applications.yml"

while getopts d OPT; do
  # shellcheck disable=SC2220
  case ${OPT} in
  "d") F_OPTION="" ;;
  esac
done

# Remove processed options from $@
shift $(($OPTIND - 1))

docker compose -p cqrs-es-example-go -f ../docker-compose/docker-compose-databases.yml ${F_OPTION} down -v --remove-orphans
docker compose -p cqrs-es-example-go -f ../docker-compose/docker-compose-databases.yml ${F_OPTION} up --remove-orphans --force-recreate --renew-anon-volumes -d "$@"
