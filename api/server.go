package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/interfaces"
	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/gorilla/mux"
)

//Server is struct for service api
type Server struct {
	port           string
	mux            *mux.Router
	routersAuth    *mux.Router
	routersNotAuth *mux.Router
	logger         interfaces.ServiceLogger
	Dynamo         *infra.DynamodbClient
	Sqs            *infra.SqsClient
}

//NewServer return new service api
func NewServer(port string) *Server {
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		log.Fatal(err)
	}
	sqs := infra.NewSqsClient()
	if err := sqs.Setup(); err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()
	authRouters := router.PathPrefix("/").Subrouter()
	notAuthRouters := router.PathPrefix("/").Subrouter()
	authRouters.Use(utils.LoggingMiddleware, utils.SetupHeadersMiddleware, utils.AuthorizeMiddleware)
	notAuthRouters.Use(utils.LoggingMiddleware, utils.SetupHeadersMiddleware)
	return &Server{
		mux:            router,
		routersAuth:    authRouters,
		routersNotAuth: notAuthRouters,
		logger:         utils.NewLogger(os.Stdout),
		port:           port,
		Dynamo:         dynamo,
		Sqs:            sqs,
	}
}

//Run execute router and listen the server
func (s *Server) Run() error {
	s.logger.Info(fmt.Sprintf("Listen in port %v", s.port))
	if err := s.setup(); err != nil {
		return err
	}
	return nil
}

func (s *Server) setup() error {
	if err := http.ListenAndServe(fmt.Sprintf(":%v", s.port), s.mux); err != nil {
		return err
	}
	return nil
}

//Health is function the health server
func (s *Server) Health() {
	s.routersNotAuth.HandleFunc("/health", s.health).Methods(http.MethodGet)
}

//CreateSubscription is function the creation subscriptions
func (s *Server) CreateSubscription() {
	s.routersAuth.HandleFunc("/subscriptions", s.createSubscription).Methods(http.MethodPost)
}

//ListSubscription return all subscriptions the client
func (s *Server) ListSubscriptions() {
	s.routersAuth.HandleFunc("/subscriptions", s.listSubscriptions).Methods(http.MethodGet)
}

//DeleteSubscription remove all or single event the subscription
func (s *Server) DeleteSubscription() {
	s.routersAuth.HandleFunc("/subscriptions/{id}", s.deleteSubscription).Methods(http.MethodDelete)
}

//TestNotification execute the send message to queue
func (s *Server) TestNotification() {
	s.routersAuth.HandleFunc("/subscriptions/test", s.testNotification).Methods(http.MethodPost)
}

//AllRouters execute all routers the server
func (s *Server) AllRouters() {
	s.routersNotAuth.HandleFunc("/health", s.health).Methods(http.MethodGet)
	s.routersAuth.HandleFunc("/subscriptions", s.listSubscriptions).Methods(http.MethodGet)
	s.routersAuth.HandleFunc("/subscriptions", s.createSubscription).Methods(http.MethodPost)
	s.routersAuth.HandleFunc("/subscriptions/{id}", s.deleteSubscription).Methods(http.MethodDelete)
	s.routersAuth.HandleFunc("/subscriptions/test", s.testNotification).Methods(http.MethodPost)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "ok"})
	if err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Server) createSubscription(w http.ResponseWriter, r *http.Request) {
	subs := dto.SubscriptionDTO{
		ClientId: r.Header.Get("client-id"),
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&subs); err != nil {
		s.logger.Error(err.Error())
		return
	}
	result, err := s.Dynamo.CreateSubscription(&subs)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	if err := json.NewEncoder(w).Encode(&result); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("client-id")
	resultSubs, err := s.Dynamo.ListSubscriptions(clientId)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	if err := json.NewEncoder(w).Encode(&resultSubs); err != nil {
		s.logger.Error(err.Error())
		return
	}

}

func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	clientId := r.Header.Get("client-id")
	if clientId == "" {
		if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "clientId is required"}); err != nil {
			s.logger.Error(err.Error())
			return
		}
	}
	event := r.URL.Query().Get("event")
	if event != "" {
		subscription, err := s.Dynamo.GetSubscription(clientId, event)
		if err != nil {
			s.logger.Error(err.Error())
		}
		if subscription.ClientId == clientId && subscription.SubscriptionEvent == event {
			if err := s.Dynamo.DeleteSubscription(clientId, id, event); err != nil {
				s.logger.Error(err.Error())
			}
		}
	} else {
		subscriptions, err := s.Dynamo.ListSubscriptions(clientId)
		if err != nil {
			s.logger.Error(err.Error())
		}
		if len(subscriptions) == 0 {
			s.logger.Info(fmt.Sprintf("not subscriptions for client_id [%s] and subscription_id [%s]\n", clientId, id))
		}
		for _, sub := range subscriptions {
			if sub.SubscriptionId == id && sub.ClientId == clientId {
				for _, ev := range sub.Events {
					if err := s.Dynamo.DeleteSubscription(clientId, id, ev); err != nil {
						s.logger.Error(err.Error())
					}
				}
			}
		}
	}
	if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "success"}); err != nil {
		s.logger.Error(err.Error())
		return
	}

}

func (s *Server) testNotification(w http.ResponseWriter, r *http.Request) {
	notif := dto.NotifierDTO{
		ClientId: r.Header.Get("client-id"),
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		s.logger.Error(err.Error())
		return
	}
	respNotSubscriptions := dto.ResponseNotifyDTO{
		Ok:     false,
		Topic:  notif.Event,
		SentTo: []dto.SubscriptionDTO{},
	}
	subscription, err := s.Dynamo.GetSubscription(notif.ClientId, notif.Event)
	if err != nil && err.Error() == "not subscription event in table" {
		if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
			s.logger.Error(err.Error())
			return
		}
		return
	}

	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	if subscription.SubscriptionEvent == notif.Event && subscription.ClientId == notif.ClientId {
		subs := dto.SubscriptionDTO{
			ClientId:          subscription.ClientId,
			AuthProvider:      "",
			Url:               subscription.SubscriptionUrl,
			SubscriptionEvent: subscription.SubscriptionEvent,
			CreatedAt:         notif.CreatedAt,
		}
		_, err := s.Sqs.Send(subs, notif)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
		responseNotify := dto.ResponseNotifyDTO{
			Ok:    true,
			Topic: subscription.SubscriptionEvent,
			SentTo: []dto.SubscriptionDTO{
				{
					ClientId:          subscription.ClientId,
					SubscriptionId:    subscription.SubscriptionId,
					SubscriptionEvent: subscription.SubscriptionEvent,
					SubscriptionUrl:   subscription.SubscriptionUrl,
				},
			},
		}

		if err := json.NewEncoder(w).Encode(&responseNotify); err != nil {
			s.logger.Error(err.Error())
			return
		}
		return
	} else {
		if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
			s.logger.Error(err.Error())
			return
		}
	}

}
