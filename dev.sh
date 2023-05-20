#!/bin/bash

set -eo pipefail

docker-compose start

export DEV=true
export TABLE_NAME=abborre
export BASE_URL=http://localhost:3000
export ADMIN_PASSWORD=admin
export SECRETS=unsafe-dev-secret
if [ -f .env ]; then
  source .env
fi

gin --port 3000 --bin abborre --all -- -dev
