package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/badico-cloud-hub/log-driver/logger"
	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/entity"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/middlewares"
	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/gorilla/mux"
)

//Server is struct for service api
type Server struct {
	port               string
	mux                *mux.Router
	routersAdmin       *mux.Router
	routersAuth        *mux.Router
	routersServiceAuth *mux.Router
	routersNotAuth     *mux.Router
	LogManager         *producer.LoggerManager
	Dynamo             *infra.DynamodbClient
	Sqs                *infra.SqsClient
	RabbitMQClient     *infra.RabbitMQ
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

	rabbitMqClient := infra.NewRabbitMQ()
	if err := rabbitMqClient.Setup(); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	adminRouters := router.PathPrefix("/").Subrouter()
	authRouters := router.PathPrefix("/").Subrouter()
	authServiceRouters := router.PathPrefix("/").Subrouter()
	notAuthRouters := router.PathPrefix("/").Subrouter()
	adminRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware, middlewares.AuthorizeAdminMiddleware)
	authRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware, middlewares.AuthorizeMiddleware)
	authServiceRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware, middlewares.AuthorizeMiddlewareByServiceApiKey)
	notAuthRouters.Use(middlewares.LoggingMiddleware, middlewares.SetupHeadersMiddleware)
	return &Server{
		mux:                router,
		routersAdmin:       adminRouters,
		routersAuth:        authRouters,
		routersServiceAuth: authServiceRouters,
		routersNotAuth:     notAuthRouters,
		port:               port,
		Dynamo:             dynamo,
		Sqs:                sqs,
		RabbitMQClient:     rabbitMqClient,
	}
}

//Run execute router and listen the server
func (s *Server) Run() error {
	runLog := s.LogManager.NewLogger("logger run application - ", os.Getenv("MACHINE_IP"))
	runLog.Infof("Listen in port %v", s.port)
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
func (s *Server) Notification() {
	s.routersServiceAuth.HandleFunc("/notify", s.notification).Methods(http.MethodPost)
}

//TestNotification by client execute the send message to queue
func (s *Server) TestNotification() {
	s.routersAuth.HandleFunc("/notify/test", s.testNotification).Methods(http.MethodPost)
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
	s.routersAdmin.HandleFunc("/services", s.createServices).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/services", s.listServices).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/services/{service}", s.getServices).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/services/{service}", s.deleteServices).Methods(http.MethodDelete)
	s.routersAdmin.HandleFunc("/services/{service}/events", s.addEvent).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/services/{service}/events", s.getServicesEvents).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/clients", s.createClients).Methods(http.MethodPost)
	s.routersAdmin.HandleFunc("/clients", s.listClients).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/clients/{service}", s.getClients).Methods(http.MethodGet)
	s.routersAdmin.HandleFunc("/clients/{service}/{association_id}", s.deleteClients).Methods(http.MethodDelete)
	s.routersAuth.HandleFunc("/subscriptions", s.createSubscription).Methods(http.MethodPost)
	s.routersAuth.HandleFunc("/subscriptions", s.listSubscriptions).Methods(http.MethodGet)
	s.routersAuth.HandleFunc("/subscriptions/{id}", s.deleteSubscription).Methods(http.MethodDelete)
	s.routersServiceAuth.HandleFunc("/notify", s.notification).Methods(http.MethodPost)
	s.routersAuth.HandleFunc("/notify/test", s.testNotification).Methods(http.MethodPost)

}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	healthLog := s.LogManager.NewLogger("logger health route - ", os.Getenv("MACHINE_IP"))

	err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "ok"})
	if err != nil {
		healthLog.Errorln(err.Error())
	}
}

func (s *Server) createSubscription(w http.ResponseWriter, r *http.Request) {
	createSubscriptionLog := s.LogManager.NewLogger("logger create subscription route - ", os.Getenv("MACHINE_IP"))
	subs := dto.SubscriptionDTO{
		ClientId:      r.Header.Get("client-id"),
		AssociationId: r.Header.Get("association-id"),
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&subs); err != nil {
		createSubscriptionLog.Errorln(err.Error())
		return
	}

	client, _ := s.Dynamo.GetClientByApiKey(r.Header.Get("c-token"))
	for _, event := range subs.Events {
		subscriptionService := strings.Split(event, ".")[0]
		_, err := s.Dynamo.GetServicesEvents(subscriptionService, event)
		if err != nil {
			createSubscriptionLog.Errorln(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
			return
		}
		if client.Service != subscriptionService {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "Not Authorized"})
			return
		}
	}

	result, err := s.Dynamo.CreateSubscription(&subs)
	if err != nil {
		createSubscriptionLog.Errorln(err.Error())
		return
	}
	jsonToLog, _ := json.Marshal(result)
	createSubscriptionLog.AddEvent(logger.LogEventEmbed{
		Name: "CreateSubscription",
		Type: "Default",
		Params: []logger.LogEventParam{
			{Value: string(jsonToLog), Key: "result"},
		},
	})
	createSubscriptionLog.Infof("Create Subscription: %+v", result)
	if err := json.NewEncoder(w).Encode(&result); err != nil {
		createSubscriptionLog.Errorln(err.Error())
		return
	}
}

func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	listSubscriptionsLog := s.LogManager.NewLogger("logger list subscriptions route - ", os.Getenv("MACHINE_IP"))
	associationId := r.Header.Get("association-id")
	listSubscriptionsLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", associationId))
	resultSubs, err := s.Dynamo.ListSubscriptions(associationId)
	if err != nil {
		listSubscriptionsLog.Errorln(err.Error())
		return
	}
	listSubscriptionsLog.Infof("Subscriptions: %+v", resultSubs)
	if err := json.NewEncoder(w).Encode(&resultSubs); err != nil {
		listSubscriptionsLog.Errorln(err.Error())
		return
	}

}

func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	deleteSubscriptionLog := s.LogManager.NewLogger("logger delete subscription route - ", os.Getenv("MACHINE_IP"))
	id := mux.Vars(r)["id"]
	associationId := r.Header.Get("association-id")
	clientId := r.Header.Get("client-id")
	deleteSubscriptionLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", associationId))
	deleteSubscriptionLog.AddTraceRef(fmt.Sprintf("ClientId: %s", clientId))
	if associationId == "" {
		if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: "associationId is required"}); err != nil {
			deleteSubscriptionLog.Errorln(err.Error())
			return
		}
	}
	event := r.URL.Query().Get("event")
	if event != "" {
		deleteSubscriptionLog.AddTraceRef(fmt.Sprintf("Event: %s", event))
		subscription, err := s.Dynamo.GetSubscription(associationId, event)
		if err != nil {
			deleteSubscriptionLog.Errorln(err.Error())
		}
		if subscription.ClientId == clientId && subscription.SubscriptionEvent == event {
			if err := s.Dynamo.DeleteSubscription(clientId, event, subscription.SubscriptionUrl); err != nil {
				deleteSubscriptionLog.Errorln(err.Error())
			}
			deleteSubscriptionLog.Infof("Subscription with AssociationId %s, Event %s, ClientId %s and Id %s deleted!", associationId, event, clientId, id)
		}
	} else {
		subscriptions, err := s.Dynamo.ListSubscriptions(associationId)
		if err != nil {
			deleteSubscriptionLog.Errorln(err.Error())
		}
		if len(subscriptions) == 0 {
			deleteSubscriptionLog.Infof(fmt.Sprintf("not subscriptions for associationId [%s] and subscription_id [%s]\n", associationId, id))
		}
		for _, sub := range subscriptions {
			if sub.SubscriptionId == id && sub.ClientId == clientId {
				for _, ev := range sub.Events {
					if err := s.Dynamo.DeleteSubscription(clientId, ev, sub.SubscriptionUrl); err != nil {
						deleteSubscriptionLog.Errorln(err.Error())
					}
				}
			}
		}
		deleteSubscriptionLog.Infof("Subscription with AssociationId %s, ClientId %s and Id %s deleted!", associationId, clientId, id)
	}
	if err := json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "success"}); err != nil {
		deleteSubscriptionLog.Errorln(err.Error())
		return
	}

}

func (s *Server) createServices(w http.ResponseWriter, r *http.Request) {
	createServiceLog := s.LogManager.NewLogger("logger create service route - ", os.Getenv("MACHINE_IP"))
	serviceDto := dto.ServicesDTO{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&serviceDto); err != nil {
		createServiceLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	createServiceLog.AddTraceRef(fmt.Sprintf("Service name: %s", serviceDto.Name))
	existService, _, err := s.Dynamo.ExistService(serviceDto.Name)
	if err != nil {
		createServiceLog.Errorln(err.Error())
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
	apiKey, err := s.Dynamo.CreateServices(serviceDto)
	if err != nil {
		createServiceLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	jsonToLog, _ := json.Marshal(serviceDto)
	createServiceLog.AddEvent(logger.LogEventEmbed{
		Name: "CreateService",
		Type: "Default",
		Params: []logger.LogEventParam{
			{Value: string(jsonToLog), Key: "serviceDto"},
		},
	})
	createServiceLog.Infof("Created service with: %+v", serviceDto)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dto.ServicesDTO{ApiKey: apiKey}); err != nil {
		createServiceLog.Errorln(err.Error())
		return
	}

}
func (s *Server) getServices(w http.ResponseWriter, r *http.Request) {
	getServicesLog := s.LogManager.NewLogger("logger get service route - ", os.Getenv("MACHINE_IP"))
	service := mux.Vars(r)["service"]
	getServicesLog.AddTraceRef(fmt.Sprintf("Service: %s", service))
	serviceEvents, err := s.Dynamo.GetServices(service)
	if err != nil {
		getServicesLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	serviceDto := dto.ServicesDTO{}
	if len(serviceEvents) > 0 {
		serviceDto.Name = serviceEvents[0].Name
		serviceDto.ServiceId = serviceEvents[0].ServiceId
		serviceDto.Entity = serviceEvents[0].Entity
		serviceDto.CreatedAt = serviceEvents[0].CreatedAt
		for _, event := range serviceEvents {
			serviceDto.Events = append(serviceDto.Events, event.ServiceEvent)
		}

		getServicesLog.Infof("Service: %+v", serviceDto)
		if err := json.NewEncoder(w).Encode(&serviceDto); err != nil {
			getServicesLog.Errorln(err.Error())
			return
		}

	} else {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()}); err != nil {
			getServicesLog.Errorln(err.Error())
			return
		}
	}

}

func (s *Server) getServicesEvents(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	event := r.URL.Query().Get("name")
	getServiceEventsLog := s.LogManager.NewLogger("logger get service events route - ", os.Getenv("MACHINE_IP"))
	getServiceEventsLog.AddTraceRef(fmt.Sprintf("Service: %s", service))
	getServiceEventsLog.AddTraceRef(fmt.Sprintf("Event: %s", event))
	serviceEvent, err := s.Dynamo.GetServicesEvents(service, event)
	if err != nil {
		getServiceEventsLog.Errorln(err.Error())
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
	getServiceEventsLog.Infof("Service Event: %+v", serviceEventDTO)
	if err := json.NewEncoder(w).Encode(&serviceEventDTO); err != nil {
		getServiceEventsLog.Errorln(err.Error())
		return
	}

}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	listServiceLog := s.LogManager.NewLogger("logger list services route - ", os.Getenv("MACHINE_IP"))
	services, err := s.Dynamo.ListServices()
	if err != nil {
		listServiceLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}

	allServices := []dto.ServicesDTO{}
	mapServices := make(map[string]dto.ServicesDTO)
	for _, serv := range services {
		serviceDto := dto.ServicesDTO{
			Name:      serv.Name,
			Entity:    serv.Entity,
			ServiceId: serv.ServiceId,
			CreatedAt: serv.CreatedAt,
		}
		mapServices[serv.Name] = serviceDto
	}
	for _, val := range mapServices {
		allServices = append(allServices, val)
	}
	listServiceLog.Infof("Services: %+v", allServices)
	if err := json.NewEncoder(w).Encode(&allServices); err != nil {
		listServiceLog.Errorln(err.Error())
		return
	}

}

func (s *Server) deleteServices(w http.ResponseWriter, r *http.Request) {
	deleteServiceLog := s.LogManager.NewLogger("logger delete service route - ", os.Getenv("MACHINE_IP"))
	service := mux.Vars(r)["service"]
	eventParam := r.URL.Query().Get("event")
	deleteServiceLog.AddTraceRef(fmt.Sprintf("Service: %s", service))
	servicesEvents, err := s.Dynamo.GetServices(service)
	if err != nil {
		deleteServiceLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}

	if eventParam != "" {
		deleteServiceLog.AddTraceRef(fmt.Sprintf("EventParam: %s", eventParam))
		deleteServiceLog.Infof("Delete Event Service: %+v", eventParam)
		for _, ev := range servicesEvents {
			if eventParam == ev.ServiceEvent {
				if err := s.Dynamo.DeleteServices(ev.Name, eventParam); err != nil {
					deleteServiceLog.Errorln(err.Error())
				}
				break
			}
		}
	} else {
		deleteServiceLog.Infof("Delete Service: %+v", service)
		for _, ev := range servicesEvents {
			if err := s.Dynamo.DeleteServices(ev.Name, ev.ServiceEvent); err != nil {
				deleteServiceLog.Errorln(err.Error())
				continue
			}
		}
	}
	if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
		deleteServiceLog.Errorln(err.Error())
		return
	}
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {
	addEventServiceLog := s.LogManager.NewLogger("logger add event in service route - ", os.Getenv("MACHINE_IP"))
	service := mux.Vars(r)["service"]
	addEventServiceLog.AddTraceRef(fmt.Sprintf("Service: %s", service))
	defer r.Body.Close()
	payload := dto.ServicesDTO{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		addEventServiceLog.Errorln(err.Error())
		return
	}

	exists, serviceId, err := s.Dynamo.ExistService(service)
	if err != nil {
		addEventServiceLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if exists {
		addEventServiceLog.AddTraceRef(fmt.Sprintf("ServiceId: %s", serviceId))
		for _, event := range payload.Events {
			existedEvent, err := s.Dynamo.GetServicesEvents(service, event)
			if err != nil && err != infra.ErrorServiceEventNotFound {
				addEventServiceLog.Errorln(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
				return
			}

			if existedEvent.ServiceEvent == event {
				addEventServiceLog.Errorln(infra.ErrorServiceEventAlreadyExist.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceEventAlreadyExist.Error()})
				return
			}

			_, err = s.Dynamo.PutEventService(service, serviceId, event)

			if err != nil {
				addEventServiceLog.Errorln(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
				return
			}
			jsonToLog, _ := json.Marshal(payload)
			addEventServiceLog.AddEvent(logger.LogEventEmbed{
				Name: "AddEvent",
				Type: "Default",
				Params: []logger.LogEventParam{
					{Value: string(jsonToLog), Key: "serviceDto"},
				},
			})
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
			addEventServiceLog.Errorln(err.Error())
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()}); err != nil {
			addEventServiceLog.Errorln(err.Error())
			return
		}

	}

}

func (s *Server) createClients(w http.ResponseWriter, r *http.Request) {
	createClientLog := s.LogManager.NewLogger("logger create client route - ", os.Getenv("MACHINE_IP"))
	client := dto.ClientDTO{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		createClientLog.Errorln(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	_, exists, err := s.Dynamo.ExistClient(client.Identifier, client.Service)
	if err != nil && err != infra.ErrorClientNotFound {
		createClientLog.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if exists {
		createClientLog.Errorln(infra.ErrorClientAlreadyExist.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorClientAlreadyExist.Error()})
		return
	}
	existService, _, err := s.Dynamo.ExistService(client.Service)
	if err != nil {
		createClientLog.Errorln(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if !existService {
		createClientLog.Errorln(infra.ErrorServiceNotFound.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()})
		return
	}
	apiKey, err := s.Dynamo.CreateClients(client)
	jsonToLog, _ := json.Marshal(client)
	createClientLog.AddEvent(logger.LogEventEmbed{
		Name: "CreateClient",
		Type: "Default",
		Params: []logger.LogEventParam{
			{Value: string(jsonToLog), Key: "clientDto"},
		},
	})
	createClientLog.Infof("Create Client: %+v", client)
	if err != nil {
		createClientLog.Errorln(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if err := json.NewEncoder(w).Encode(&dto.ClientDTO{ApiKey: apiKey}); err != nil {
		createClientLog.Errorln(err.Error())
		return
	}
}

func (s *Server) getClients(w http.ResponseWriter, r *http.Request) {
	listClientsServiceLog := s.LogManager.NewLogger("logger list clients service route - ", os.Getenv("MACHINE_IP"))
	service := mux.Vars(r)["service"]
	identifier := r.URL.Query().Get("identifier")
	listClientsServiceLog.AddTraceRef(fmt.Sprintf("Service: %s", service))
	defer r.Body.Close()

	exist, _, err := s.Dynamo.ExistService(service)
	if err != nil {
		listClientsServiceLog.Errorln(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	if !exist {
		listClientsServiceLog.Errorln(infra.ErrorServiceNotFound.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: infra.ErrorServiceNotFound.Error()})
		return
	}
	clients, err := s.Dynamo.ListClients()
	if err != nil {
		listClientsServiceLog.Errorln(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	responseClients := []dto.ClientDTO{}
	if len(identifier) > 0 {
		for _, client := range clients {
			if client.Service == service && client.Identifier == identifier {
				newClient := dto.ClientDTO{
					ApiKey:        client.ApiKey,
					Identifier:    client.Identifier,
					Service:       client.Service,
					AssociationId: client.AssociationId,
					CreatedAt:     client.CreatedAt,
					UpdatedAt:     client.UpdatedAt,
				}
				responseClients = append(responseClients, newClient)
			}
		}
	} else {
		for _, client := range clients {
			if client.Service == service {
				newClient := dto.ClientDTO{
					ApiKey:        client.ApiKey,
					Identifier:    client.Identifier,
					Service:       client.Service,
					AssociationId: client.AssociationId,
					CreatedAt:     client.CreatedAt,
					UpdatedAt:     client.UpdatedAt,
				}
				responseClients = append(responseClients, newClient)
			}
		}

	}
	listClientsServiceLog.Infof("Response Clients: %+v", responseClients)
	if err := json.NewEncoder(w).Encode(&responseClients); err != nil {
		listClientsServiceLog.Errorln(err.Error())
		return
	}
}
func (s *Server) listClients(w http.ResponseWriter, r *http.Request) {
	listClientsLog := s.LogManager.NewLogger("logger list clients route - ", os.Getenv("MACHINE_IP"))
	clients, err := s.Dynamo.ListClients()
	if err != nil {
		listClientsLog.Errorln(err.Error())
		_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
		return
	}
	responseClients := []dto.ClientDTO{}
	for _, client := range clients {
		respClient := dto.ClientDTO{
			ApiKey:        client.ApiKey,
			Identifier:    client.Identifier,
			Service:       client.Service,
			AssociationId: client.AssociationId,
			CreatedAt:     client.CreatedAt,
			UpdatedAt:     client.UpdatedAt,
		}
		responseClients = append(responseClients, respClient)
	}
	listClientsLog.Infof("List clients: %+v", responseClients)
	if err := json.NewEncoder(w).Encode(&responseClients); err != nil {
		listClientsLog.Errorln(err.Error())
		return
	}
}
func (s *Server) deleteClients(w http.ResponseWriter, r *http.Request) {
	deleteClientLog := s.LogManager.NewLogger("logger delete client route - ", os.Getenv("MACHINE_IP"))
	service := mux.Vars(r)["service"]
	association_id := mux.Vars(r)["association_id"]
	deleteClientLog.AddTraceRef(fmt.Sprintf("Service: %s", service))
	deleteClientLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", association_id))

	client, exists, err := s.Dynamo.ExistClient(association_id, service)

	if err != nil {
		deleteClientLog.Errorln(err.Error())
		if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
			deleteClientLog.Errorln(err.Error())
		}
		return
	}

	apiKey := strings.Split(client.PK, "#")[1]
	if exists {
		if err := s.Dynamo.DeleteClients(apiKey, service); err != nil {
			deleteClientLog.Errorln(err.Error())
			_ = json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "error", Message: err.Error()})
			return
		}
	}
	deleteClientLog.Infof("Client with AssociationId %s deleted!", association_id)
	if err := json.NewEncoder(w).Encode(&dto.ResponseDTO{Status: "success"}); err != nil {
		deleteClientLog.Errorln(err.Error())
		return
	}

}

func (s *Server) notification(w http.ResponseWriter, r *http.Request) {
	notificationLog := s.LogManager.NewLogger("logger notification route - ", os.Getenv("MACHINE_IP"))
	notif := dto.NotifierDTO{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		notificationLog.Errorln(err.Error())
		return
	}
	notificationLog.AddTraceRef(fmt.Sprintf("ClientId: %s", notif.ClientId))

	respNotSubscriptions := dto.ResponseNotifyDTO{
		Ok:     false,
		Topic:  notif.Event,
		SentTo: []dto.SubscriptionDTO{},
	}

	subscriptions := []entity.Subscription{}
	subscriptionsFiltered := []dto.SubscriptionDTO{}

	for _, associationId := range notif.AssociationsId {
		newSubsciptions, err := s.Dynamo.GetSubscriptionByAssociationIdAndEvent(associationId, notif.Event)
		if err != nil && err == infra.ErrorSubscriptinEventNotFound {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
				notificationLog.Errorln(err.Error())
				return
			}
			return
		}

		if err != nil {
			notificationLog.Errorln(err.Error())
			return
		}

		subscriptions = append(subscriptions, newSubsciptions...)
	}

	for _, subscription := range subscriptions {
		if duplicated := utils.VerifyIfUrlIsDuplicated(subscriptionsFiltered, subscription.SubscriptionUrl, subscription.SubscriptionEvent); !duplicated {
			subscriptionsFiltered = append(subscriptionsFiltered, dto.SubscriptionDTO{
				AssociationId:     subscription.AssociationId,
				ClientId:          subscription.ClientId,
				AuthProvider:      "",
				Url:               subscription.SubscriptionUrl,
				SubscriptionEvent: subscription.SubscriptionEvent,
				CreatedAt:         notif.CreatedAt,
			})
		}
	}

	for _, subscription := range subscriptionsFiltered {
		notif.Data["topic"] = subscription.SubscriptionEvent
		notif.Data["created_at"] = subscription.CreatedAt
		queueMessage := dto.QueueMessage{
			ClientId:      subscription.ClientId,
			AssociationId: subscription.AssociationId,
			Url:           subscription.SubscriptionUrl,
			AuthProvider:  subscription.AuthProvider,
			Body:          notif.Data,
		}
		err := s.RabbitMQClient.Producer(queueMessage)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
				notificationLog.Errorln(err.Error())
				return
			}
			notificationLog.Errorln(err.Error())
			return
		}
	}

	respNotSubscriptions = dto.ResponseNotifyDTO{
		Ok:     true,
		Topic:  notif.Event,
		SentTo: subscriptionsFiltered,
	}

	notificationLog.Infof("Send notification to Subscriptions: %+v", respNotSubscriptions)

	if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
		notificationLog.Errorln(err.Error())
		return
	}
	return

}

func (s *Server) testNotification(w http.ResponseWriter, r *http.Request) {
	testNotificationLog := s.LogManager.NewLogger("logger test notification route - ", os.Getenv("MACHINE_IP"))
	notif := dto.NotifierDTO{
		ClientId:       r.Header.Get("client-id"),
		AssociationsId: []string{r.Header.Get("association-id")},
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		testNotificationLog.Errorln(err.Error())
		return
	}
	testNotificationLog.AddTraceRef(fmt.Sprintf("Event: %s", notif.Event))
	testNotificationLog.AddTraceRef(fmt.Sprintf("ClientId: %s", notif.ClientId))

	respNotSubscriptions := dto.ResponseNotifyDTO{
		Ok:     false,
		Topic:  notif.Event,
		SentTo: []dto.SubscriptionDTO{},
	}

	subscriptionsFiltered := []dto.SubscriptionDTO{}
	associationId := notif.AssociationsId[0]
	subscriptions, err := s.Dynamo.GetSubscriptionByAssociationIdAndEvent(associationId, notif.Event)
	testNotificationLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", associationId))

	if err != nil && err == infra.ErrorSubscriptinEventNotFound {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
			testNotificationLog.Errorln(err.Error())
			return
		}
		return
	}

	if err != nil {
		testNotificationLog.Errorln(err.Error())
		return
	}

	for _, subscription := range subscriptions {
		if duplicated := utils.VerifyIfUrlIsDuplicated(subscriptionsFiltered, subscription.SubscriptionUrl, subscription.SubscriptionEvent); !duplicated {
			subscriptionsFiltered = append(subscriptionsFiltered, dto.SubscriptionDTO{
				AssociationId:     associationId,
				ClientId:          subscription.ClientId,
				AuthProvider:      "",
				SubscriptionUrl:   subscription.SubscriptionUrl,
				SubscriptionEvent: subscription.SubscriptionEvent,
				CreatedAt:         notif.CreatedAt,
			})
		}
	}

	for _, subscription := range subscriptionsFiltered {
		notif.Data["topic"] = subscription.SubscriptionEvent
		notif.Data["created_at"] = subscription.CreatedAt
		queueMessage := dto.QueueMessage{
			ClientId:      subscription.ClientId,
			AssociationId: subscription.AssociationId,
			Url:           subscription.SubscriptionUrl,
			AuthProvider:  subscription.AuthProvider,
			Body:          notif.Data,
		}
		err := s.RabbitMQClient.Producer(queueMessage)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
				testNotificationLog.Errorln(err.Error())
				return
			}
			testNotificationLog.Errorln(err.Error())
			return
		}
	}

	subscriptionsResponse := []dto.SubscriptionDTO{}

	for _, subs := range subscriptionsFiltered {
		subscriptionsResponse = append(subscriptionsResponse, dto.SubscriptionDTO{
			SubscriptionUrl: subs.SubscriptionUrl,
		})
	}

	respNotSubscriptions = dto.ResponseNotifyDTO{
		Ok:     true,
		Topic:  notif.Event,
		SentTo: subscriptionsResponse,
	}

	testNotificationLog.Infof("Send notification to Subscriptions: %+v", respNotSubscriptions)

	if err := json.NewEncoder(w).Encode(&respNotSubscriptions); err != nil {
		testNotificationLog.Errorln(err.Error())
		return
	}
	return

}
