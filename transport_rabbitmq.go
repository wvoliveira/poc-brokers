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
	s.RabbitServer()
}

func (s service) RabbitServer() {
	go s.RabbitFindServer()
}

func (s service) RabbitFind(id string) (payload []byte, err error) {
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
		log.Fatalf("Failed to register a consumer: %s", err.Error())
	}

	correlationID := uuid.NewString()

	// We set the payload for the message.
	body := id
	err = ch.PublishWithContext(
		context.TODO(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		rabbitmq.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       q.Name,
			Body:          []byte(body),
		})

	if err != nil {
		log.Fatalf("Failed to publish a message: %s", err.Error())
	}

	for d := range msgs {
		if correlationID == d.CorrelationId {
			payload = d.Body
			break
		}
	}

	fmt.Println("PAYLOAD")
	fmt.Println(string(payload))
	return
}

func (s service) RabbitFindServer() {
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
		log.Fatalf("Failed to declare a queue: %s\n", err.Error())
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %s\n", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s\n", err.Error())
	}

	forever := make(chan bool)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for msg := range msgs {
			id := string(msg.Body)
			log.Printf("ID: %s\n", id)

			var (
				response = Response{
					Status: "successful",
				}
				item Item
				body []byte
			)

			if id != "" {
				item, err = s.Find(id)
				if err != nil {
					log.Printf("Failed to find ID: %s\n", err.Error())
					response.Status = "error"
					response.Message = ErrNotFound.Error()
				}
			}

			if item.ID != "" && response.Status != "error" {
				body, err = json.Marshal(item)
				if err != nil {
					log.Printf("Failed to marshal item struct: %s\n", err.Error())
					response.Status = "error"
					response.Message = ErrMarshalJson.Error()
				} else {
					response.Data = body
				}
			}

			body, err = json.Marshal(response)
			if err != nil {
				log.Printf("Failed to marshal response struct: %s\n", err.Error())
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

			fmt.Println("Body:")
			fmt.Println(string(body))

			msg.Ack(false)
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
