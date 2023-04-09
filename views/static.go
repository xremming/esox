package views

import (
	"embed"
	"net/http"
)

//go:embed static/*
var static embed.FS

func Static() http.Handler {
	return http.FileServer(http.FS(static))
}
