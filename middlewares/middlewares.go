package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/helpers"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/utils"
)

// LoggingMiddleware is middleware to parse logs
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggerMiddleware := utils.NewLogger(os.Stdout)
		loggerMiddleware.Info(fmt.Sprintf("%s %s", r.Method, r.URL))
		next.ServeHTTP(w, r)
	})
}

// SetupHeadersMiddleware is middleware to set Content-Type
func SetupHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// AuthorizeMiddleware is middleware to authorize routers
func AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := utils.NewLogger(os.Stdout)
		dynamo := infra.NewDynamodbClient()
		if err := dynamo.Setup(); err != nil {
			log.Fatal(err)
		}
		adminApiKey := r.Header.Get("a-token")
		bearerToken := r.Header.Get("authorization")

		if bearerToken != "" {
			jwtDto, err := helpers.VerifyToken(bearerToken)
			if err != nil {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusForbidden)
				if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
					logger.Error(err.Error())
				}
				return
			}
			r.Header.Add("client-id", jwtDto.Payload.ClientId)
			r.Header.Add("association-id", jwtDto.Payload.AssociationId)
			r.Header.Add("api-key-type", jwtDto.Payload.ApiKeyType)
			r.Header.Add("scopes", jwtDto.Payload.Scopes)
			next.ServeHTTP(w, r)
			return
		}
		if adminApiKey != "" {
			secretManager := infra.NewSecretManagerClient()
			adminBase64 := secretManager.Get("ADMIN_APIS_KEY").(string)
			isValid, admin := utils.ValidateAdminApiKey(adminApiKey, adminBase64)
			if isValid {
				clientId := r.Header.Get("client-id")
				if clientId != "" {
					client, err := dynamo.GetClientsByClientId(clientId)
					fmt.Printf("client: %+v", client)
					fmt.Printf("err: %+v", err)
					if err != nil {
						w.WriteHeader(http.StatusForbidden)
						if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
							logger.Error(err.Error())
						}
						return
					}
					r.Header.Add("client-id", client.Identifier)
					r.Header.Add("association-id", client.AssociationId)
					r.Header.Add("api-key-type", client.Service)
					r.Header.Add("provider", client.Provider)

				}
				r.Header.Add("admin-client-id", admin.ClientId)
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
				logger.Error(err.Error())
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
			logger.Error(err.Error())
			return
		}
	})
}

// AuthorizeMiddleware is middleware to authorize routers by service API KEY
func AuthorizeMiddlewareByServiceApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := utils.NewLogger(os.Stdout)
		dynamo := infra.NewDynamodbClient()
		if err := dynamo.Setup(); err != nil {
			log.Fatal(err)
		}
		headerApiKey := r.Header.Get("s-token")
		if headerApiKey != "" {
			_, err := dynamo.GetServiceByApiKey(headerApiKey)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
					logger.Error(err.Error())
				}
				return
			}
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
				logger.Error(err.Error())
				return
			}
		}
	})
}

// AuthorizeAdminMiddleware is middleware to authorize admin routers
func AuthorizeAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := utils.NewLogger(os.Stdout)
		headerApiKey := r.Header.Get("a-token")
		strBase64 := os.Getenv("ADMIN_APIS_KEY")
		adminApisKeysJson, err := base64.StdEncoding.DecodeString(strBase64)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
				logger.Error(err.Error())
			}
			return
		}

		adminObjects := []dto.AdminObject{}

		if err := json.Unmarshal(adminApisKeysJson, &adminObjects); err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
				logger.Error(err.Error())
			}
			return
		}

		for _, admin := range adminObjects {
			if admin.ApiKey == headerApiKey {
				r.Header.Add("client-id", admin.ClientId)
				next.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "unalthorized"}); err != nil {
			logger.Error(err.Error())
		}
	})
}
