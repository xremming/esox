package esox

import (
	"net/http"

	"github.com/justinas/alice"
)

func notFoundMiddleware(notFound http.Handler) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				notFound.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
