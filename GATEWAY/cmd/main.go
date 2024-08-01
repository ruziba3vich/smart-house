package main

import (
	"context"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ruziba3vich/smart-house/app"
	"github.com/ruziba3vich/smart-house/app/handler"

	usersprotos "github.com/ruziba3vich/smart-house/genprotos/submodules/users_submodule/protos"
	"github.com/ruziba3vich/smart-house/internal/config"
	"github.com/ruziba3vich/smart-house/internal/msgbroker"
	"github.com/ruziba3vich/smart-house/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	conn, err := amqp.Dial(config.GetRabbitMqURI())
	if err != nil {
		logger.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer conn.Close()

	gconn, err := grpc.NewClient("localhost:7000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("Error opening channel: %v", err)
	}
	defer ch.Close()

	registration, rq, err := getMessages("create", ch)
	if err != nil {
		logger.Fatalf("Error getting registration messages: %v", err)
	}

	updates, uq, err := getMessages("update", ch)
	if err != nil {
		logger.Fatalf("Error getting update messages: %v", err)
	}

	deletions, dq, err := getMessages("delete", ch)
	if err != nil {
		logger.Fatalf("Error getting deletion messages: %v", err)
	}

	msgBroker, err := msgbroker.NewRPCClient(ch, 10*time.Second, registration, updates, deletions, ctx)
	if err != nil {
		logger.Fatalf("Error creating RPC client: %v", err)
	}

	app := app.New(
		handler.NewRbmqHandler(logger, msgBroker, utils.NewTokenGenerator(config), config, rq, uq, dq),
		handler.NewGrpcHandler(logger, usersprotos.NewUsersServiceClient(gconn)),
	)
	if err := app.RUN(config); err != nil {
		logger.Fatalf("Application error: %v", err)
	}
}

func getMessages(queueName string, ch *amqp.Channel) (<-chan amqp.Delivery, amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, q, err
	}

	messages, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	return messages, q, err
}
