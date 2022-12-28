package middlewares

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/utils"
)

//LoggingMiddleware is middleware to parse logs
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggerMiddleware := utils.NewLogger(os.Stdout)
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
		logger := utils.NewLogger(os.Stdout)
		dynamo := infra.NewDynamodbClient()
		if err := dynamo.Setup(); err != nil {
			log.Fatal(err)
		}
		headerApiKey := r.Header.Get("api-key")
		if headerApiKey != "" {
			client, err := dynamo.GetClientByApiKey(headerApiKey)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "Not Authorized"}); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			r.Header.Add("client-id", client.Identifier)
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "Not Authorized"}); err != nil {
				logger.Error(err.Error())
				return
			}
		}
	})
}
