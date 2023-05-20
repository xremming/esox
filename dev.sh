#!/bin/bash

set -eo pipefail

docker-compose start
gin --port 3000 --bin abborre --all -- -dev
