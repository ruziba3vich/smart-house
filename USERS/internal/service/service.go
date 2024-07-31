package service

import (
	"context"
	"fmt"
	"log"

	genprotos "github.com/ruziba3vich/users/genprotos/users_submodule/protos"
	"github.com/ruziba3vich/users/internal/redisservice"
	"github.com/ruziba3vich/users/internal/storage"
)

type (
	Service struct {
		storage *storage.Storage
		redis   *redisservice.RedisService
		logger  *log.Logger
		genprotos.UnimplementedUsersServiceServer
	}
)

func New(storage *storage.Storage, redis *redisservice.RedisService, logger *log.Logger) *Service {
	return &Service{
		storage: storage,
		redis:   redis,
		logger:  logger,
	}
}

func (s *Service) RegisterUser(ctx context.Context, req *genprotos.CreateUserReuest) (*genprotos.Response, error) {
	s.logger.Println("-- RECEIVED A REQUEST TO <RegisterUser> SERVICE --")
	user, err := s.storage.CreateUser(ctx, req)
	var response genprotos.Response
	if err == nil {
		if err := s.redis.StoreUserInRedis(ctx, user); err != nil {
			return nil, err
		}
		response.Message = "user has successfully been registered"
		return &response, nil
	}
	response.Message = "failed to register user"
	return &response, err
}

func (s *Service) LoginUser(ctx context.Context, req *genprotos.LoginRequest) (*genprotos.RegisterUserResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <LoginUser> SERVICE --")
	response, err := s.storage.LoginUser(ctx, req)
	if err == nil {
		s.redis.StoreUserInRedis(ctx, response.User)
	}
	return response, err
}

func (s *Service) GetById(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.User, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <GetById> SERVICE --")
	user, err := s.redis.GetUserFromRedis(ctx, req.GetByField)
	if err == nil {
		if user != nil {
			return user, nil
		}
	}
	return s.storage.GetUserById(ctx, req)
}

func (s *Service) GetByUsername(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.User, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <GetByUsername> SERVICE --")
	return s.storage.GetUserByUsername(ctx, req)
}

func (s *Service) GetByEmail(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.User, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <GetByEmail> SERVICE --")
	return s.storage.GetUserByEmail(ctx, req)
}

func (s *Service) UpdateUser(ctx context.Context, req *genprotos.UpdateUserReuqest) (*genprotos.Response, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <UpdateUser> SERVICE --")
	updatedUser, err := s.storage.UpdateUser(ctx, req)
	var response genprotos.Response
	if err == nil {
		if err := s.redis.StoreUserInRedis(ctx, updatedUser); err != nil {
			return nil, err
		}
		response.Message = "user has successfully been updated"
	} else {
		response.Message = fmt.Sprintf("user could not be updated : %s", err.Error())
	}
	return &response, err
}

func (s *Service) GetAllUsers(ctx context.Context, req *genprotos.GetAllUsersRequest) (*genprotos.GetAllUsersResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <GetAllUsers> SERVICE --")
	return s.storage.GetAllUsers(ctx, req)
}

func (s *Service) DeleteUserById(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.Response, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <DeleteUserById> SERVICE --")
	if err := s.storage.DeleteUserById(ctx, req); err != nil {
		return nil, err
	}
	if err := s.redis.DeleteUserFromRedis(ctx, req.GetByField); err != nil {
		return nil, err
	}

	return &genprotos.Response{
		Message: "user has successfully been deleted",
	}, nil
}

func (s *Service) GetUsersByAddress(ctx context.Context, req *genprotos.GetUsersByAddressRequest) (*genprotos.GetAllUsersResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST IN <GetUsersByAddress> SERVICE --")
	return s.storage.GetUserByAddress(ctx, req)
}

/*
   rpc RegisterUser(CreateUserReuest) returns (Response); /// ------------
   rpc LoginUser(LoginRequest) returns (RegisterUserResponse); -----------
   rpc GetById(GetByFieldRequest) returns (User); ----------
   rpc GetByUsername(GetByFieldRequest) returns (User); -----------
   rpc GetByEmail(GetByFieldRequest) returns (User); ------------
   rpc UpdateUser(UpdateUserReuqest) returns (Response); /// --------
   rpc GetAllUsers(google.protobuf.Empty) returns (GetAllUsersResponse); ----------
   rpc DeleteUserById(GetByFieldRequest) returns (Response); /// -----------
   rpc GetUsersByAddress(GetUsersByAddressRequest) returns (GetAllUsersResponse);
*/
