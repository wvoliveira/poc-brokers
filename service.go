package main

import (
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type service struct {
	rabbit *rabbitmq.Connection
}

func NewService(rabbit *rabbitmq.Connection) service {
	return service{
		rabbit: rabbit,
	}
}

func (s service) Find(id string) (item Item, err error) {
	var found bool

	for _, data := range Data {
		if id == data.ID {
			found = true
			item = data
			break
		}
	}

	if !found {
		err = ErrNotFound
	}
	return
}

func (s service) FindAll() (items []Item) {
	return Data
}
