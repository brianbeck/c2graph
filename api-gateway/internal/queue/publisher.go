package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// ScanMessage is the message published to the scan_requests queue.
type ScanMessage struct {
	JobID        string `json:"job_id"`
	Address      string `json:"address"`
	Depth        int    `json:"depth"`
	CurrentDepth int    `json:"current_depth"`
	RootJobID    string `json:"root_job_id"`
}

// Publisher handles publishing messages to RabbitMQ.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher creates a new RabbitMQ publisher with retry.
func NewPublisher(uri string) (*Publisher, error) {
	var conn *amqp.Connection
	var err error

	for attempt := 0; attempt < 10; attempt++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			break
		}
		log.Warn().Err(err).Int("attempt", attempt+1).Msg("Failed to connect to RabbitMQ, retrying...")
		time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Ensure queue exists
	_, err = ch.QueueDeclare("scan_requests", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

// Publish sends a scan message to the queue.
func (p *Publisher) Publish(ctx context.Context, msg *ScanMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		"",
		"scan_requests",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// Close cleanly shuts down the publisher.
func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
