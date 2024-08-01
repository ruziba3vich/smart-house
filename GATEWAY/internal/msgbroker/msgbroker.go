package msgbroker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MsgBroker struct {
	ch           *amqp.Channel
	timeoutCh    <-chan time.Time
	Registration <-chan amqp.Delivery
	Updates      <-chan amqp.Delivery
	Deletes      <-chan amqp.Delivery
	ctx          context.Context
}

func NewRPCClient(ch *amqp.Channel, timeout time.Duration, Registration <-chan amqp.Delivery, Updates <-chan amqp.Delivery, Deletes <-chan amqp.Delivery, ctx context.Context) (*MsgBroker, error) {
	timeoutCh := time.After(timeout)
	return &MsgBroker{
		ch:           ch,
		timeoutCh:    timeoutCh,
		Registration: Registration,
		Updates:      Updates,
		Deletes:      Deletes,
		ctx:          ctx,
	}, nil
}

func (m *MsgBroker) PublishToQueue(messages <-chan amqp.Delivery, body []byte, q amqp.Queue, replyToQueue, contentType string) ([]byte, error) {
	corrId := uuid.New().String()
	fmt.Println("Generated a new CorrelationId:", corrId)

	err := m.ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:   contentType,
			CorrelationId: corrId,
			ReplyTo:       replyToQueue,
			Body:          body,
		},
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case d := <-messages:
			fmt.Println("Received message with CorrelationId:", d.CorrelationId)
			if corrId == d.CorrelationId {
				return d.Body, nil
			}
		case <-m.timeoutCh:
			return nil, context.DeadlineExceeded
		case <-m.ctx.Done():
			return nil, m.ctx.Err()
		}
	}
}
