package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmqConn, err := rabbitmq.Dial("amqp://guest:guest@server:5672/")
	if err != nil {
		log.Fatalf("Failed to open a connection with RabbitMQ: %s", err.Error())
	}

	defer rabbitmqConn.Close()
	router := mux.NewRouter()

	{
		s := NewService(rabbitmqConn)
		s.NewHTTP(router)
		s.NewRabbit()
	}

	addr := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("HTTP handler http://%s\n", addr)
	log.Fatal(srv.ListenAndServe())
}
