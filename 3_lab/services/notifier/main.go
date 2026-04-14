package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string `envconfig:"RABBIT_HOST" default:"localhost"`
	Port     string `envconfig:"RABBIT_PORT" default:"5672"`
	Username string `envconfig:"RABBIT_USERNAME" default:"guest"`
	Password string `envconfig:"RABBIT_PASSWORD" default:"guest"`
}

func (c Config) URL() string {
	return "amqp://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/"
}

const (
	exchange   = "notes.events"
	queue      = "notifier.queue"
	routingKey = "note.*"
)

type Event struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	// Retry connection loop
	var conn *amqp.Connection
	var err error
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(cfg.URL())
		if err == nil {
			break
		}
		log.Printf("waiting for RabbitMQ... attempt %d/10", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Declare same exchange as producer
	err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Declare queue
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Bind queue to exchange with wildcard routing key
	err = ch.QueueBind(q.Name, routingKey, exchange, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[notifier] listening on exchange=%s queue=%s routingKey=%s", exchange, queue, routingKey)

	for msg := range msgs {
		var event Event
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[notifier] failed to parse message: %v", err)
			msg.Nack(false, false)
			continue
		}
		log.Printf("[notifier] received event: type=%s payload=%s timestamp=%s",
			event.Type, string(event.Payload), event.Timestamp.Format(time.RFC3339))
		msg.Ack(false)
	}
}
