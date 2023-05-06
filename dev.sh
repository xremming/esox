#!/bin/bash

docker-compose start
gin --port 3000 --bin abborre-local --build cmd/local --all -- -dev
