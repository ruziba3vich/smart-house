package msgbroker

import (
	"context"
	"encoding/json"
	"log"
	controlrpc "ruziba3vich/github.com/control/genprotos/controller_submodule"
	"ruziba3vich/github.com/control/internal/storage"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MsgBrokerService struct {
	msgs           <-chan amqp.Delivery
	storageService *storage.Storage
	logger         *log.Logger
}

func NewService(msgs <-chan amqp.Delivery, storageService *storage.Storage, logger *log.Logger) *MsgBrokerService {
	return &MsgBrokerService{
		msgs: msgs,
		storageService: storageService,
		logger: logger,
	}
}

func (m *MsgBrokerService) ConsumeMessages(ch *amqp.Channel, queueName string, logger *log.Logger, handler func(context.Context, *storage.Storage, *amqp.Delivery, *log.Logger)) {
	for msg := range m.msgs {
		handler(context.Background(), m.storageService, &msg, logger)
	}
}

func (m *MsgBrokerService) HandleTurnDeviceOn(ctx context.Context, msg *amqp.Delivery) {
	var req controlrpc.DeviceRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		m.logger.Printf("Failed to unmarshal message: %v", err)
		return
	}

	_, err := m.storageService.TurnDeviceOn(ctx, &req)
	if err != nil {
		m.logger.Printf("Failed to turn device on: %v", err)
		return
	}
	m.logger.Printf("Device %s turned on successfully", req.DeviceId)
}

func (m *MsgBrokerService) HandleTurnDeviceOff(ctx context.Context, msg *amqp.Delivery) {
	var req controlrpc.DeviceRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		m.logger.Printf("Failed to unmarshal message: %v", err)
		return
	}

	_, err := m.storageService.TurnDeviceOff(ctx, &req)
	if err != nil {
		m.logger.Printf("Failed to turn device off: %v", err)
		return
	}
	m.logger.Printf("Device %s turned off successfully", req.DeviceId)
}

func (m *MsgBrokerService) HandleAddUserToHouse(ctx context.Context, msg *amqp.Delivery) {
	var req controlrpc.UserRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		m.logger.Printf("Failed to unmarshal message: %v", err)
		return
	}

	_, err := m.storageService.AddUserToHouse(ctx, &req)
	if err != nil {
		m.logger.Printf("Failed to add user to house: %v", err)
		return
	}
	m.logger.Printf("User %s added to house %s successfully", req.UserId, req.HouseId)
}

func (m *MsgBrokerService) HandleRemoveUserFromHouse(ctx context.Context, msg *amqp.Delivery) {
	var req controlrpc.UserRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		m.logger.Printf("Failed to unmarshal message: %v", err)
		return
	}

	_, err := m.storageService.RemoveUserFromHouse(ctx, &req)
	if err != nil {
		m.logger.Printf("Failed to remove user from house: %v", err)
		return
	}
	m.logger.Printf("User %s removed from house %s successfully", req.UserId, req.HouseId)
}
