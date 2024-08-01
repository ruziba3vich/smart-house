package service

import (
	"context"
	"log"

	genprotos "github.com/ruziba3vich/devices/genprotos/devices_submodule"
	"github.com/ruziba3vich/devices/internal/redisservice"
	"github.com/ruziba3vich/devices/internal/storage"
)

type (
	Service struct {
		storage *storage.Storage
		redis   *redisservice.RedisService
		logger  *log.Logger
		genprotos.UnimplementedDeviceServiceServer
	}
)

func New(storage *storage.Storage, redis *redisservice.RedisService, logger *log.Logger) *Service {
	return &Service{
		storage: storage,
		redis:   redis,
		logger:  logger,
	}
}

func (s *Service) CreateDevice(ctx context.Context, req *genprotos.CreateDeviceRequest) (*genprotos.CreateDeviceResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST TO <CreateDevice> SERVICE --")
	device, err := s.storage.CreateDevice(ctx, req)
	var response genprotos.CreateDeviceResponse
	if err == nil {
		if err := s.redis.StoreDeviceInRedis(ctx, device.Device); err != nil {
			return nil, err
		}
		response.Device = device.Device
		return &response, nil
	}
	response.Device = nil
	return &response, err
}

func (s *Service) UpdateDevice(ctx context.Context, req *genprotos.UpdateDeviceRequest) (*genprotos.UpdateDeviceResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST TO <UpdateDevice> SERVICE --")
	updatedDevice, err := s.storage.UpdateDevice(ctx, req)
	var response genprotos.UpdateDeviceResponse
	if err == nil {
		if err := s.redis.StoreDeviceInRedis(ctx, updatedDevice.Device); err != nil {
			return nil, err
		}
		response.Device = updatedDevice.Device
	} else {
		response.Device = nil
	}
	return &response, err
}

func (s *Service) GetDevice(ctx context.Context, req *genprotos.GetDeviceRequest) (*genprotos.GetDeviceResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST TO <GetDevice> SERVICE --")
	device, err := s.redis.GetDeviceFromRedis(ctx, req.Id)
	if err == nil {
		if device != nil {
			return device, nil
		}
	}

	device, err = s.storage.GetDevice(ctx, req)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s *Service) DeleteDevice(ctx context.Context, req *genprotos.DeleteDeviceRequest) (*genprotos.DeleteDeviceResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST TO <DeleteDevice> SERVICE --")
	response, err := s.storage.DeleteDevice(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.redis.DeleteDeviceFromRedis(ctx, req.Id); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) GetAllDevices(ctx context.Context, req *genprotos.GetAllDevicesRequest) (*genprotos.GetAllDevicesResponse, error) {
	s.logger.Println("-- RECEIVED A REQUEST TO <GetAllDevices> SERVICE --")
	devices, err := s.storage.GetAllDevices(ctx, req)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
