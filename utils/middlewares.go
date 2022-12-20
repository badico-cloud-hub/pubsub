package utils

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

//LoggingMiddleware is middleware to parse logs
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggerMiddleware := NewLogger(os.Stdout)
		loggerMiddleware.Info(fmt.Sprintf("%s %s", r.Method, r.URL))
		next.ServeHTTP(w, r)
	})
}

//SetupHeadersMiddleware is middleware to set Content-Type
func SetupHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

//AuthorizeMiddleware is middleware to authorize routers
func AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerApiKey := r.Header.Get("api-key")
		if headerApiKey != "" && strings.Contains(headerApiKey, ".") {
			subListApiKey := strings.Split(headerApiKey, ".")
			clientId := subListApiKey[0]
			r.Header.Add("client-id", clientId)
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
