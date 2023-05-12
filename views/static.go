package views

import (
	"net/http"
	"os"
)

var static = os.DirFS(".")

func Static() http.Handler {
	return http.FileServer(http.FS(static))
}
