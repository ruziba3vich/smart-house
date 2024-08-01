package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ruziba3vich/github.com/control/internal/config"
	"ruziba3vich/github.com/control/internal/models"
	"ruziba3vich/github.com/control/internal/msgbroker"
	"ruziba3vich/github.com/control/internal/storage"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	logger := log.New(log.Writer(), "Service: ", log.LstdFlags)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	db, err := storage.ConnectDB(cfg, ctx)
	if err != nil {
		logger.Fatal(err)
	}
	storageService := storage.NewStorage(db, logger)

	conn, err := amqp.Dial(cfg.GetRabbitMqURI())
	if err != nil {
		logger.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queues := []models.TYPE{
		models.TURNDEVICEONQUEUE,
		models.TURNDEVICEOFFQUEUE,
		models.ADDUSERQUEUE,
		models.REMOVEUSERQUEUE,
	}
	for _, q := range queues {
		_, err := ch.QueueDeclare(
			string(q),
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			logger.Fatalf("Failed to declare a queue: %v", err)
		}
	}

	msgs := make(chan amqp.Delivery)
	msgBrokerService := msgbroker.NewService(msgs, storageService, logger)

	go FunctionToRunConsumer(ch, models.TURNDEVICEONQUEUE, logger, msgBrokerService, msgBrokerService.HandleTurnDeviceOn)
	go FunctionToRunConsumer(ch, models.TURNDEVICEOFFQUEUE, logger, msgBrokerService, msgBrokerService.HandleTurnDeviceOff)
	go FunctionToRunConsumer(ch, models.ADDUSERQUEUE, logger, msgBrokerService, msgBrokerService.HandleAddUserToHouse)
	go FunctionToRunConsumer(ch, models.REMOVEUSERQUEUE, logger, msgBrokerService, msgBrokerService.HandleRemoveUserFromHouse)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	logger.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(2 * time.Second)
}

func FunctionToRunConsumer(ch *amqp.Channel, queueName models.TYPE, logger *log.Logger, msgBrokerService *msgbroker.MsgBrokerService, handler func(context.Context, *amqp.Delivery)) {
	msgs, err := ch.Consume(
		string(queueName),
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatalf("Failed to register a consumer: %v", err)
	}
	go msgBrokerService.ConsumeMessages(ch, string(queueName), logger, handler)
}
