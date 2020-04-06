package internal

import (
	"fmt"
	"github.com/abdelhamidbakhta/senkyou/internal/broker"
	"github.com/abdelhamidbakhta/senkyou/internal/log"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

var (
	logger = log.ForceGetLogger()
)

type SenkyouServer interface {
	Start()
}

func NewSenkyouServer(config Config, broker broker.Broker) SenkyouServer {
	return server{
		config: config,
		broker: broker,
	}
}

type server struct {
	config Config
	broker broker.Broker
}

func (s server) Start() {
	logger.Info("starting senkyou http server")
	fmt.Println(s.config.string())
	router := mux.NewRouter()
	router.HandleFunc("/", s.home)
	router.HandleFunc("/pub/{topic}/", s.pub)
	router.HandleFunc("/sub/{topic}/", s.sub)
	logger.Error("cannot start senkyou server", zap.Error(http.ListenAndServe(s.config.ListenAddr(), router)))
}

func (s server) pub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.broker.Publish(topic, body)
	if err != nil {
		logger.Error("failed to publish message", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s server) sub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	err := s.broker.Subscribe(topic, func(message []byte) {
		logger.Info("received message")
		fmt.Println(string(message))
	})
	if err != nil {
		logger.Error("failed to subscribe to topic", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (server) home(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("senkyou is up!\n"))
}