package middleware

import (
	"net/http"
)

const internalTokenHeader = "X-Internal-Token"

// InterService validates shared secret headers for server-to-server communication.
func InterService(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			received := r.Header.Get(internalTokenHeader)
			if received == "" || received != expectedToken {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
