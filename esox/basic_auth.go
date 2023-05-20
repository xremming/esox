package esox

import "net/http"

func BasicAuthFunc(auth func(username, password string) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || !auth(u, p) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized.", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return BasicAuthFunc(func(u, p string) bool {
		return u == username && p == password
	})
}
