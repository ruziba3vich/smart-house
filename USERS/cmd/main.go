package main

import (
	"context"
	"crypto/sha256"
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ruziba3vich/users/grpcapp"
	"github.com/ruziba3vich/users/internal/config"
	"github.com/ruziba3vich/users/internal/msgbroker"
	"github.com/ruziba3vich/users/internal/redisservice"
	"github.com/ruziba3vich/users/internal/service"
	"github.com/ruziba3vich/users/internal/storage"
)

func main() {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := context.Background()
	// defer cancel()

	db, err := storage.ConnectDB(cfg, ctx)
	if err != nil {
		logger.Fatal(err)
	}

	hash := sha256.New()

	redisService := redisservice.New(redis.NewClient(&redis.Options{
		Addr: cfg.GetRedisURI(),
		DB:   0,
	}), logger)

	service := service.New(storage.NewStorage(db, logger, hash, cfg), redisService, logger)

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

	grpcserver := grpcapp.NewUsersApp(service)

	regQueue, err := getQueue(ch, "create")
	if err != nil {
		logger.Fatal(err)
	}
	regMsgs, err := getMessageQueue(ch, regQueue)
	if err != nil {
		logger.Fatal(err)
	}

	updQueue, err := getQueue(ch, "update")
	if err != nil {
		logger.Fatal(err)
	}
	updMsgs, err := getMessageQueue(ch, updQueue)
	if err != nil {
		logger.Fatal(err)
	}

	delQueue, err := getQueue(ch, "delete")
	if err != nil {
		logger.Fatal(err)
	}
	delMsgs, err := getMessageQueue(ch, delQueue)
	if err != nil {
		logger.Fatal(err)
	}

	msgBroker := msgbroker.New(service, ch, logger, regMsgs, updMsgs, delMsgs, &sync.WaitGroup{}, 3)

	// Start gRPC server in a separate goroutine
	go func() {
		logger.Fatal(grpcserver.RUN(cfg, logger))
	}()

	msgBroker.StartToConsume(ctx, "application/json")
}

func getQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
}

func getMessageQueue(ch *amqp.Channel, q amqp.Queue) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}
