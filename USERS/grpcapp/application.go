package grpcapp

import (
	"log"
	"net"

	genprotos "github.com/ruziba3vich/users/genprotos/users_submodule/protos"
	"github.com/ruziba3vich/users/internal/config"
	"google.golang.org/grpc"
)

type (
	GRPCApp struct {
		service genprotos.UsersServiceServer
	}
)

func New(service genprotos.UsersServiceServer) *GRPCApp {
	return &GRPCApp{
		service: service,
	}
}

func (a *GRPCApp) RUN(cfg *config.Config, logger *log.Logger) error {
	listener, err := net.Listen(cfg.Protocol, cfg.Port)
	if err != nil {
		logger.Printf("ERROR WHILE CREATING A LISTENER %s\n", err.Error())
		return err
	}
	serverRegisterer := grpc.NewServer()
	genprotos.RegisterUsersServiceServer(serverRegisterer, a.service)
	logger.Printf("--- SERVER HAS STARTED TO RUN ON PORT %s\n", cfg.Port)
	return serverRegisterer.Serve(listener)
}
