package msgbroker

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	genprotos "github.com/ruziba3vich/users/genprotos/users_submodule/protos"
	"google.golang.org/protobuf/proto"
)

type (
	MsgBroker struct {
		service          genprotos.UsersServiceServer
		channel          *amqp.Channel
		registrations    <-chan amqp.Delivery
		profileUpdates   <-chan amqp.Delivery
		profileDeletions <-chan amqp.Delivery
		logger           *log.Logger
		wg               *sync.WaitGroup
		numberOfServices int
		// connection *amqp.Connection
	}
)

func New(service genprotos.UsersServiceServer,
	channel *amqp.Channel,
	logger *log.Logger,
	registrations <-chan amqp.Delivery,
	profileUpdates <-chan amqp.Delivery,
	profileDeletions <-chan amqp.Delivery,
	wg *sync.WaitGroup,
	numberOfServices int) *MsgBroker {
	return &MsgBroker{
		service:          service,
		channel:          channel,
		registrations:    registrations,
		profileUpdates:   profileUpdates,
		profileDeletions: profileDeletions,
		logger:           logger,
		wg:               wg,
		numberOfServices: numberOfServices,
	}
}

func (m *MsgBroker) StartToConsume(ctx context.Context, contentType string) {
	m.wg.Add(m.numberOfServices)
	/// registration consumer
	go func() {
		defer m.wg.Done()
		for {
			select {
			case val := <-m.registrations:
				var request genprotos.CreateUserReuest
				if err := json.Unmarshal(val.Body, &request); err != nil {
					m.logger.Printf("ERROR WHILE MARSHALING DATA : %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				response, err := m.service.RegisterUser(ctx, &request)
				if err != nil {
					m.logger.Printf("failed in registration %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				val.Ack(false)

				byteData, err := proto.Marshal(response)
				if err != nil {
					m.logger.Printf("failed to marshal data %s\n", err.Error())
					continue
				}

				err = m.channel.Publish(
					"",
					val.ReplyTo,
					false,
					false,
					amqp.Publishing{
						ContentType:   contentType,
						CorrelationId: val.CorrelationId,
						Body:          byteData,
					},
				)
				if err != nil {
					m.logger.Printf("failed to publish message : %s", err.Error())
				}
			case <-ctx.Done():
				m.logger.Println("request took too long to execute")
				return
			}
		}
	}()

	/// updates consumer
	go func() {
		defer m.wg.Done()
		for {
			select {
			case val := <-m.profileUpdates:
				var request genprotos.UpdateUserReuqest
				if err := json.Unmarshal(val.Body, &request); err != nil {
					m.logger.Printf("ERROR WHILE MARSHALING DATA : %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				response, err := m.service.UpdateUser(ctx, &request)
				if err != nil {
					m.logger.Printf("failed in registration %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				val.Ack(false)

				byteData, err := proto.Marshal(response)
				if err != nil {
					m.logger.Printf("failed to marshal data %s\n", err.Error())
					continue
				}

				err = m.channel.Publish(
					"",
					val.ReplyTo,
					false,
					false,
					amqp.Publishing{
						ContentType:   contentType,
						CorrelationId: val.CorrelationId,
						Body:          byteData,
					},
				)
				if err != nil {
					m.logger.Printf("failed to publish message : %s", err.Error())
				}
			case <-ctx.Done():
				m.logger.Println("request took too long to execute")
				return
			}
		}
	}()

	/// deletes consumer

	go func() {
		defer m.wg.Done()
		for {
			select {
			case val := <-m.profileUpdates:
				var request genprotos.GetByFieldRequest
				if err := json.Unmarshal(val.Body, &request); err != nil {
					m.logger.Printf("ERROR WHILE MARSHALING DATA : %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				response, err := m.service.DeleteUserById(ctx, &request)
				if err != nil {
					m.logger.Printf("failed in deleting %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				val.Ack(false)

				byteData, err := proto.Marshal(response)
				if err != nil {
					m.logger.Printf("failed to marshal data %s\n", err.Error())
					continue
				}

				err = m.channel.Publish(
					"",
					val.ReplyTo,
					false,
					false,
					amqp.Publishing{
						ContentType:   contentType,
						CorrelationId: val.CorrelationId,
						Body:          byteData,
					},
				)
				if err != nil {
					m.logger.Printf("failed to publish message : %s", err.Error())
				}
			case <-ctx.Done():
				m.logger.Println("request took too long to execute")
				return
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	m.wg.Wait()
	m.logger.Println("SHUT DOWN COMPLETE")
}
