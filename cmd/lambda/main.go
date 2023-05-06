package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/xremming/abborre/app"
	"github.com/xremming/abborre/esox"
)

func Handler(ctx context.Context, event *events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error) {
	log := esox.SetupLogger(false)
	ctx = log.WithContext(ctx)

	tableName, ok := os.LookupEnv("TABLE_NAME")
	if !ok {
		log.Fatal().Msg("TABLE_NAME not set")
	}
	handler := app.NewHandler(ctx, app.Configuration{IsDev: false, TableName: tableName})

	return lambdaurl.Wrap(handler)(ctx, event)
}

func main() {
	lambda.Start(Handler)
}
