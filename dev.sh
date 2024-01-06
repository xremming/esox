#!/bin/bash

set -eo pipefail

docker-compose start

export DEV=true
export TABLE_NAME=abborre
export BASE_URL=http://localhost:3000
export SECRETS=unsafe-dev-secret
if [ -f .env ]; then
  source .env
fi

go build
PORT=3000 DEV=true ./abborre
