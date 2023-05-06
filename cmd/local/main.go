package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/xremming/abborre/app"
	"github.com/xremming/abborre/esox"
)

func GetEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func GetEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		res, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}

		return res
	}

	return defaultValue
}

var (
	flagDev       = flag.Bool("dev", GetEnvBoolOrDefault("DEV", false), "Development mode")
	flagHost      = flag.String("host", GetEnvOrDefault("HOST", "localhost"), "HTTP host")
	flagPort      = flag.String("port", GetEnvOrDefault("PORT", "8000"), "HTTP port")
	flagTableName = flag.String("table-name", GetEnvOrDefault("TABLE_NAME", "abborre"), "DynamoDB table name")
)

func init() {
	flag.Parse()
}

func main() {
	log := esox.SetupLogger(*flagDev)
	ctx := log.WithContext(context.Background())

	handler := app.NewHandler(ctx, app.Configuration{IsDev: *flagDev, TableName: *flagTableName})

	addr := *flagHost + ":" + *flagPort
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Start server in goprocess.
	go func() {
		log.Info().
			Bool("dev", *flagDev).
			Str("tableName", *flagTableName).
			Str("addr", addr).
			Msg("HTTP server starting")

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("HTTP server closed")
		} else {
			log.Err(err).Msg("HTTP server ListenAndServe")
		}
	}()

	// Wait for a signal to quit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Shutdown the server.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
