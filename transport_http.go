package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s service) NewHTTP(router *mux.Router) {
	r := router.PathPrefix("/").Subrouter()

	r.HandleFunc("/http/find", s.HTTPFindAll).Methods("GET")
	r.HandleFunc("/http/find/{id}", s.HTTPFind).Methods("GET")

	r.HandleFunc("/rabbitmq/find", s.HTTPRabbitMQFindAll).Methods("GET")
	r.HandleFunc("/rabbitmq/find/{id}", s.HTTPRabbitMQFind).Methods("GET")
}

func (s service) HTTPFind(w http.ResponseWriter, r *http.Request) {
	id := DecodeFind(r)
	item, err := s.Find(id)
	SendResponse(w, item, err, false)
}

func (s service) HTTPFindAll(w http.ResponseWriter, r *http.Request) {
	_, _, _ = DecodeFindAll()
	items := s.FindAll()
	SendResponse(w, items, nil, false)
}

func (s service) HTTPRabbitMQFind(w http.ResponseWriter, r *http.Request) {
	id := DecodeFind(r)

	payload, err := s.RabbitFind(id)
	if err != nil {
		SendResponse(w, payload, err, false)
		return
	}

	response, err := EncodeFind(payload)

	fmt.Println("Response: ")
	fmt.Println(response)
	SendResponse(w, response, err, true)
}

func (s service) HTTPRabbitMQFindAll(w http.ResponseWriter, r *http.Request) {
	payload, _ := json.Marshal(Data)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(payload)
}
