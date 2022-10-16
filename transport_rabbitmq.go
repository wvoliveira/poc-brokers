package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

func (s service) NewRabbit() {
	go s.RabbitFindResponse()
}

func (s service) RabbitFindRequest(payload string, contentType string) {
	ch, err := s.rabbit.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err.Error())
	}
	defer ch.Close()

	// We create a Queue to send the message to.
	q, err := ch.QueueDeclare(
		"find", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err.Error())
	}

	if contentType == "" {
		contentType = "application/json"
	}

	correlationID := uuid.NewString()

	// We set the payload for the message.
	body := payload
	err = ch.PublishWithContext(
		context.TODO(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		rabbitmq.Publishing{
			ContentType:   contentType,
			Body:          []byte(body),
			CorrelationId: correlationID,
			ReplyTo:       q.Name,
		})

	if err != nil {
		log.Fatalf("Failed to publish a message: %s", err.Error())
	}

	log.Printf(" Congrats, sending message: %s", body)
}

func (s service) RabbitFindResponse() {
	ch, err := s.rabbit.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err.Error())
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"find", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err.Error())
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer")
	}

	forever := make(chan bool)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for msg := range msgs {
			payload := string(msg.Body)
			log.Printf("Payload: %s", string(payload))

			var (
				response Response
				item     Item
				body     []byte
			)

			// TODO: create decode.go and encode.go
			// for decode request and encode response.
			switch msg.ContentType {
			case "application/json":
				err = json.Unmarshal(msg.Body, &item)

				if err != nil {
					log.Printf("Error to decode payload as json: %s", err.Error())

					response = Response{
						Status:  "error",
						Message: "Error to decode payload as json",
					}
				}
			default:
				response = Response{
					Status:  "error",
					Message: fmt.Sprintf("Content type %s is not supported", msg.ContentType),
				}
			}

			if item.ID != "" {
				item, err = s.Find(item.ID)
				response.Data = item

				body, err = json.Marshal(response)
				if err != nil {
					log.Printf("Failed to marshal response struct: %s\n", err.Error())
				}
			}

			err = ch.PublishWithContext(ctx,
				"",          // exchange
				msg.ReplyTo, // routing key
				false,       // mandatory
				false,       // immediate
				rabbitmq.Publishing{
					ContentType:   msg.ContentType,
					CorrelationId: msg.CorrelationId,
					Body:          body,
				})

			if err != nil {
				log.Printf("Failed to publish a message: %s\n", err.Error())
			}

			msg.Ack(false)
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
