#!/bin/bash

set -eo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 dev|test|prod"
    exit 1
elif [ "$1" != "dev" ] && [ "$1" != "test" ] && [ "$1" != "prod" ]; then
    echo "Unknown environment $1"
    exit 1
fi

echo "Building..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o abborre -tags lambda.norpc
rm -f abborre-lambda.zip
zip -r abborre-lambda.zip ./abborre ./templates/ ./static/

export AWS_REGION=eu-north-1

echo "Deploying to $1..."
aws lambda update-function-code \
    --function-name "abborre-$1-homepage" \
    --zip-file fileb://abborre-lambda.zip \
    --publish \
    | cat

echo "Waiting for deployment to finish..."
aws lambda wait function-updated \
    --function-name "abborre-$1-homepage" \
    | cat
