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
	"github.com/badico-cloud-hub/pubsub/middlewares"
	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/gorilla/mux"
)

//Server is struct for service api
type Server struct {
	port           string
	mux            *mux.Router
	routersAdmin   *mux.Router
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
	adminRouters := router.PathPrefix("/").Subrouter()
	authRouters := router.PathPrefix("/").Subrouter()
	notAuthRouters := router.PathPrefix("/").Subrouter()
	adminRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware, middlewares.AuthorizeAdminMiddleware)
	authRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware, middlewares.AuthorizeMiddleware)
	notAuthRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware)
	return &Server{
		mux:            router,
		routersAdmin:   adminRouters,
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

//AddRouter add dynamic router for server
func (s *Server) AddRouter(path string, handler func(w http.ResponseWriter, r *http.Request), middlewares ...mux.MiddlewareFunc) {
	subrouter := s.mux.PathPrefix("/").Subrouter()
	for _, midd := range middlewares {
		if midd == nil {
			log.Fatalln("middleware of type <nil> not permited")
			return
		}
	}
	subrouter.Use(middlewares...)
	subrouter.HandleFunc(path, handler)
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

//CreateServices execute creation of services
func (s *Server) CreateServices() {
	s.routersAdmin.HandleFunc("/services", s.createServices).Methods(http.MethodPost)
}

//GetServices return service events
func (s *Server) GetServices() {
	s.routersAdmin.HandleFunc("/services/{service}", s.getServices).Methods(http.MethodGet)
}

//GetServicesEvents return all events from service
func (s *Server) GetServicesEvents() {
	s.routersAdmin.HandleFunc("/services/{service}/events", s.getServicesEvents).Methods(http.MethodGet)
}

//ListServices return all services the table
func (s *Server) ListServices() {
	s.routersAdmin.HandleFunc("/services", s.listServices).Methods(http.MethodGet)
}

//DeletServices remove service or single events the service
func (s *Server) DeleteServices() {
	s.routersAdmin.HandleFunc("/services/{service}", s.deleteServices).Methods(http.MethodDelete)
}

//AddEventToService execute append the new events to service
func (s *Server) AddEventToService() {
	s.routersAdmin.HandleFunc("/services/{service}/events", s.addEvent).Methods(http.MethodPost)
}

//CreateClients execute creation the clients in table
func (s *Server) CreateClients() {
	s.routersAdmin.HandleFunc("/clients", s.createClients).Methods(http.MethodPost)
}

//GetClients return all clients by service
func (s *Server) GetClients() {
	s.routersAdmin.HandleFunc("/clients/{service}", s.getClients).Methods(http.MethodGet)
}

//ListClients return all clients the table
func (s *Server) ListClients() {
	s.routersAdmin.HandleFunc("/clients", s.listClients).Methods(http.MethodGet)
}

//DeleteClients remove client the service
func (s *Server) DeleteClients() {
	s.routersAdmin.HandleFunc("/clients/{service}/{identifier}", s.deleteClients).Methods(http.MethodDelete)
}

//AllRouters execute all routers the server
func (s *Server) AllRouters() {
	s.routersNotAuth.HandleFunc("/health", s.health).Methods(http.MethodGet)
	s.routersAuth.HandleFunc("/subscriptions", s.listSubscriptions).Methods(http.MethodGet)
	s.routersAuth.HandleFunc("/subscriptions", s.createSubscription).Methods(http.MethodPost)
	s.routersAuth.HandleFunc("/subscriptions/{id}", s.deleteSubscription).Methods(http.MethodDelete)
	s.routersAuth.HandleFunc("/subscriptions/test", s.testNotification).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/services", s.listServices).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/services", s.createServices).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/services/{service}", s.getServices).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/services/{service}/events", s.getServicesEvents).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/services/{service}", s.deleteServices).Methods(http.MethodDelete)
	s.routersAdmin.HandleFunc("/services/{service}/events", s.addEvent).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/clients", s.createClients).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/clients", s.listClients).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/clients/{service}", s.getClients).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/clients/{service}/{identifier}", s.deleteClients).Methods(http.MethodDelete)

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

func (s *Server) createServices(w http.ResponseWriter, r *http.Request) {
	serviceDto := dto.ServicesDTO{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&serviceDto); err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	existService, _, err := s.Dynamo.ExistService(serviceDto.Name)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if existService {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceAlreadyExist.Error()})
		return
	}
	//TODO: Verify if array events is empty, and return error
	serviceId, err := s.Dynamo.CreateServices(serviceDto)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dto.ServicesDTO{ServiceId: serviceId}); err != nil {
		s.logger.Error(err.Error())
		return
	}

}
func (s *Server) getServices(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	serviceEvents, err := s.Dynamo.GetServices(service)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	serviceDto := dto.ServicesDTO{}
	if len(serviceEvents) > 0 {
		serviceDto.Name = serviceEvents[0].Name
		serviceDto.ServiceId = serviceEvents[0].ServiceId
		serviceDto.CreatedAt = serviceEvents[0].CreatedAt
		for _, event := range serviceEvents {
			serviceDto.Events = append(serviceDto.Events, event.ServiceEvent)
		}
		if err := json.NewEncoder(w).Encode(&serviceDto); err != nil {
			s.logger.Error(err.Error())
			return
		}

	} else {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()}); err != nil {
			s.logger.Error(err.Error())
			return
		}
	}

}

func (s *Server) getServicesEvents(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	event := r.URL.Query().Get("name")
	serviceEvent, err := s.Dynamo.GetServicesEvents(service, event)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	serviceEventDTO := dto.ServiceEventsDTO{
		Name:      serviceEvent.ServiceEvent,
		Service:   serviceEvent.Name,
		ServiceId: serviceEvent.ServiceId,
		CreatedAt: serviceEvent.CreatedAt,
		UpdatedAt: serviceEvent.UpdatedAt,
	}
	if err := json.NewEncoder(w).Encode(&serviceEventDTO); err != nil {
		s.logger.Error(err.Error())
		return
	}

}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.Dynamo.ListServices()
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}

	allServices := []dto.ServicesDTO{}
	mapServices := make(map[string]dto.ServicesDTO)
	for _, serv := range services {
		serviceDto := dto.ServicesDTO{
			Name:      serv.Name,
			ServiceId: serv.ServiceId,
			CreatedAt: serv.CreatedAt,
		}
		mapServices[serv.Name] = serviceDto
	}
	for _, val := range mapServices {
		allServices = append(allServices, val)
	}
	if err := json.NewEncoder(w).Encode(&allServices); err != nil {
		s.logger.Error(err.Error())
		return
	}

}
func (s *Server) deleteServices(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	eventParam := r.URL.Query().Get("event")
	servicesEvents, err := s.Dynamo.GetServices(service)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}

	if eventParam != "" {
		for _, ev := range servicesEvents {
			if eventParam == ev.ServiceEvent {
				if err := s.Dynamo.DeleteServices(ev.ServiceId, eventParam); err != nil {
					s.logger.Error(err.Error())
				}
				break
			}
		}
	} else {
		for _, ev := range servicesEvents {
			if err := s.Dynamo.DeleteServices(ev.ServiceId, ev.ServiceEvent); err != nil {
				s.logger.Error(err.Error())
				continue
			}
		}
	}
	if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	defer r.Body.Close()
	payload := dto.ServicesDTO{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error(err.Error())
		return
	}
	exists, serviceId, err := s.Dynamo.ExistService(service)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if exists {
		for _, event := range payload.Events {
			existedEvent, err := s.Dynamo.GetServicesEvents(service, event)
			if err != nil && err != infra.ErrorServiceEventNotFound {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
				return
			}

			if existedEvent.ServiceEvent == event {
				s.logger.Error(infra.ErrorServiceEventAlreadyExist.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceEventAlreadyExist.Error()})
				return
			}

			_, err = s.Dynamo.PutEventService(service, serviceId, event)

			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
			s.logger.Error(err.Error())
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()}); err != nil {
			s.logger.Error(err.Error())
			return
		}

	}

}
func (s *Server) createClients(w http.ResponseWriter, r *http.Request) {
	client := dto.ClientDTO{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	_, exists, err := s.Dynamo.ExistClient(client.Identifier, client.Service)
	if err != nil && err != infra.ErrorClientNotFound {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if exists {
		s.logger.Error(infra.ErrorClientAlreadyExist.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorClientAlreadyExist.Error()})
		return
	}
	existService, _, err := s.Dynamo.ExistService(client.Service)
	if err != nil {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if !existService {
		s.logger.Error(infra.ErrorServiceNotFound.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()})
		return
	}
	apiKey, err := s.Dynamo.CreateClients(client)
	if err != nil {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if err := json.NewEncoder(w).Encode(&dto.ClientDTO{ApiKey: apiKey}); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *Server) getClients(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	identifier := r.URL.Query().Get("identifier")
	defer r.Body.Close()

	exist, _, err := s.Dynamo.ExistService(service)
	if err != nil {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if !exist {
		s.logger.Error(infra.ErrorServiceNotFound.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()})
		return
	}
	clients, err := s.Dynamo.ListClients()
	if err != nil {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	responseClients := []dto.ClientDTO{}
	if len(identifier) > 0 {
		for _, client := range clients {
			if client.Service == service && client.Identifier == identifier {
				newClient := dto.ClientDTO{
					ApiKey:     client.GSIPK,
					Identifier: client.Identifier,
					Service:    client.Service,
					CreatedAt:  client.CreatedAt,
					UpdatedAt:  client.UpdatedAt,
				}
				responseClients = append(responseClients, newClient)
			}
		}
	} else {
		for _, client := range clients {
			if client.Service == service {
				newClient := dto.ClientDTO{
					ApiKey:     client.GSIPK,
					Identifier: client.Identifier,
					Service:    client.Service,
					CreatedAt:  client.CreatedAt,
					UpdatedAt:  client.UpdatedAt,
				}
				responseClients = append(responseClients, newClient)
			}
		}

	}
	if err := json.NewEncoder(w).Encode(&responseClients); err != nil {
		s.logger.Error(err.Error())
		return
	}
}
func (s *Server) listClients(w http.ResponseWriter, r *http.Request) {
	clients, err := s.Dynamo.ListClients()
	if err != nil {
		s.logger.Error(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	responseClients := []dto.ClientDTO{}
	for _, client := range clients {
		respClient := dto.ClientDTO{
			ApiKey:     client.GSIPK,
			Identifier: client.Identifier,
			Service:    client.Service,
			CreatedAt:  client.CreatedAt,
			UpdatedAt:  client.UpdatedAt,
		}
		responseClients = append(responseClients, respClient)
	}

	if err := json.NewEncoder(w).Encode(&responseClients); err != nil {
		s.logger.Error(err.Error())
		return
	}
}
func (s *Server) deleteClients(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	identifier := mux.Vars(r)["identifier"]

	client, exists, err := s.Dynamo.ExistClient(identifier, service)

	if err != nil {
		s.logger.Error(err.Error())
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
			s.logger.Error(err.Error())
		}
		return
	}

	if exists {
		if err := s.Dynamo.DeleteClients(client.GSIPK, service); err != nil {
			s.logger.Error(err.Error())
			_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
			return
		}
	}

	if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
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
	if err != nil && err == infra.ErrorSubscriptinEventNotFound {
		w.WriteHeader(http.StatusNotFound)
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
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
			s.logger.Error(err.Error())
			return
		}
	}

}
