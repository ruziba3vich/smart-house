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
	genprotos "github.com/ruziba3vich/devices/genprotos/devices_submodule"
	"github.com/ruziba3vich/devices/internal/models"
	"google.golang.org/protobuf/proto"
)

type (
	MsgBroker struct {
		service          genprotos.DeviceServiceServer
		channel          *amqp.Channel
		deviceCreations  <-chan amqp.Delivery
		deviceUpdates    <-chan amqp.Delivery
		deviceDeletions  <-chan amqp.Delivery
		logger           *log.Logger
		wg               *sync.WaitGroup
		numberOfServices int
	}
)

func New(service genprotos.DeviceServiceServer,
	channel *amqp.Channel,
	logger *log.Logger,
	deviceCreations <-chan amqp.Delivery,
	deviceUpdates <-chan amqp.Delivery,
	deviceDeletions <-chan amqp.Delivery,
	wg *sync.WaitGroup,
	numberOfServices int) *MsgBroker {
	return &MsgBroker{
		service:          service,
		channel:          channel,
		deviceCreations:  deviceCreations,
		deviceUpdates:    deviceUpdates,
		deviceDeletions:  deviceDeletions,
		logger:           logger,
		wg:               wg,
		numberOfServices: numberOfServices,
	}
}

func (m *MsgBroker) StartToConsume(ctx context.Context, contentType string) {
	m.wg.Add(m.numberOfServices)
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go m.consumeMessages(consumerCtx, m.deviceCreations, m.service.CreateDevice, "creation")
	go m.consumeMessages(consumerCtx, m.deviceUpdates, m.service.UpdateDevice, "update")
	go m.consumeMessages(consumerCtx, m.deviceDeletions, m.service.DeleteDevice, "deletion")

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
			case "creation":
				var req models.Device
				if err := json.Unmarshal(val.Body, &req); err != nil {
					m.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				request = req.ToCreateDeviceRequest()
				response, err = serviceFunc.(func(context.Context, *genprotos.CreateDeviceRequest) (*genprotos.CreateDeviceResponse, error))(ctx, request.(*genprotos.CreateDeviceRequest))
			case "update":
				var req models.Device
				if err := json.Unmarshal(val.Body, &req); err != nil {
					m.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				request = req.ToUpdateDeviceRequest()
				response, err = serviceFunc.(func(context.Context, *genprotos.UpdateDeviceRequest) (*genprotos.UpdateDeviceResponse, error))(ctx, request.(*genprotos.UpdateDeviceRequest))
			case "deletion":
				var req models.DeleteDeviceRequest
				if err := json.Unmarshal(val.Body, &req); err != nil {
					m.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s\n", err.Error())
					val.Nack(false, false)
					continue
				}
				request = &genprotos.GetDeviceRequest{Id: req.DeviceId}
				response, err = serviceFunc.(func(context.Context, *genprotos.GetDeviceRequest) (*genprotos.DeleteDeviceResponse, error))(ctx, request.(*genprotos.GetDeviceRequest))
			}

			if err != nil {
				m.logger.Printf("Failed in %s: %s\n", logPrefix, err.Error())
				val.Nack(false, false)
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
