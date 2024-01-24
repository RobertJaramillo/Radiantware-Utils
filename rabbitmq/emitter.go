package Radiantware_Rabbitmq

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	emitter := Emitter{
		connection: conn,
	}
	err := emitter.setup()
	if err != nil {
		return Emitter{}, err
	}

	return emitter, nil
}

func (emitter *Emitter) setup() error {

	channel, err := emitter.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return DeclareExchange(channel)

}

func (emitter *Emitter) Push(event string, level string) error {

	channel, err := emitter.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	log.Println("Pushing event to channel")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		ctx,          // Context used to communicate with exchange
		"logs_topic", // Exchange name
		level,        // Type of message we are logging
		false,        // Mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		return err
	}

	return nil

}
