package Radiantware_Rabbitmq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return consumer, nil
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return DeclareExchange(channel)
}

func (consumer *Consumer) Listen(topics []string) error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	queue, err := DeclareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, val := range topics {

		//Bind each topic to a queue
		err := channel.QueueBind(
			queue.Name,
			val,
			"logs_topic",
			false,
			nil,
		)
		if err != nil {
			return err
		}

	}

	// Consume the messages in our queue
	messages, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	//Now make a channel that will run a go routine forever to process our messages
	foreverChannel := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go HandlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\r\n", queue.Name)
	<-foreverChannel

	return nil
}

func HandlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := LogEvent(payload)
		if err != nil {
			log.Println(err)
		}
	default:
		err := LogEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func LogEvent(entry Payload) error {

	jsonData, _ := json.Marshal(entry)
	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer((jsonData)))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
