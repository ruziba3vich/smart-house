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
	"github.com/ruziba3vich/users/internal/models"
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
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go m.consumeMessages(consumerCtx, m.registrations, m.service.RegisterUser, "registration")
	go m.consumeMessages(consumerCtx, m.profileUpdates, m.service.UpdateUser, "update")
	go m.consumeMessages(consumerCtx, m.profileDeletions, m.service.DeleteUserById, "deletion")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	m.logger.Println("Shutting down, waiting for consumers to finish")
	cancel()
	m.wg.Wait()
	m.logger.Println("All consumers have stopped")
}

func (m *MsgBroker) consumeMessages(ctx context.Context, messages <-chan amqp.Delivery, serviceFunc interface{}, logPrefix string) {
	defer m.wg.Done()
	for {
		select {
		case val := <-messages:
			var request interface{}
			var response proto.Message
			var err error

			switch logPrefix {
			case "registration":
				var req models.User
				if err := json.Unmarshal(val.Body, &req); err != nil {
					m.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())
					val.Nack(false, false)
					// m.publishMessageBack(val, contentType, []byte(fmt.Sprintf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())))
					continue
				}
				request = req.ToCreateUserRequest()
				response, err = serviceFunc.(func(context.Context, *genprotos.CreateUserReuest) (*genprotos.Response, error))(ctx, request.(*genprotos.CreateUserReuest))
			case "update":
				var req models.User
				if err := json.Unmarshal(val.Body, &req); err != nil {
					m.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())
					val.Nack(false, false)
					// m.publishMessageBack(val, contentType, []byte(fmt.Sprintf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())))
					continue
				}
				request = req.ToUpdateUserRequest()
				response, err = serviceFunc.(func(context.Context, *genprotos.UpdateUserReuqest) (*genprotos.Response, error))(ctx, request.(*genprotos.UpdateUserReuqest))
			case "deletion":
				var req models.DeleteUserRequest
				if err := json.Unmarshal(val.Body, &req); err != nil {
					m.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())
					val.Nack(false, false)
					// m.publishMessageBack(val, contentType, []byte(fmt.Sprintf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())))
					continue
				}
				request = &genprotos.GetByFieldRequest{GetByField: req.UserId}
				response, err = serviceFunc.(func(context.Context, *genprotos.GetByFieldRequest) (*genprotos.Response, error))(ctx, request.(*genprotos.GetByFieldRequest))
			}

			if err != nil {
				m.logger.Printf("Failed in %s: %s\n", logPrefix, err.Error())
				val.Nack(false, false)
				// m.publishMessageBack(val, contentType, []byte(fmt.Sprintf("Failed in %s: %s\n", logPrefix, err.Error())))
				continue
			}

			val.Ack(false)

			_, err = proto.Marshal(response)
			if err != nil {
				m.logger.Printf("Failed to marshal response: %s\n", err.Error())
				continue
			}

			// m.publishMessageBack(val, contentType, byteData)
		case <-ctx.Done():
			m.logger.Printf("Context done, stopping %s consumer", logPrefix)
			return
		}
	}
}
