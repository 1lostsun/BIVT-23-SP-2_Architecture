package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "notes.events"

type Event struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	// Declare topic exchange (durable)
	err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	event := Event{
		Type:      routingKey,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	err = p.channel.PublishWithContext(ctx, exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}
	log.Printf("[publisher] sent event: routingKey=%s", routingKey)
	return nil
}

func (p *Publisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
