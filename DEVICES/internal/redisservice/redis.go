package redisservice

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	genprotos "github.com/ruziba3vich/devices/genprotos/devices_submodule"
)

type (
	RedisService struct {
		redisDb *redis.Client
		logger  *log.Logger
	}
)

func New(redisDb *redis.Client, logger *log.Logger) *RedisService {
	return &RedisService{
		logger:  logger,
		redisDb: redisDb,
	}
}

func (r *RedisService) StoreDeviceInRedis(ctx context.Context, device *genprotos.Device) error {
	deviceJSON, err := json.Marshal(device)
	if err != nil {
		r.logger.Printf("ERROR WHILE MARSHALING DATA: %s", err.Error())
		return err
	}

	err = r.redisDb.Set(ctx, device.Id, deviceJSON, time.Hour*24).Err()
	if err != nil {
		r.logger.Printf("ERROR WHILE STORING DATA IN REDIS: %s", err.Error())
		return err
	}
	return nil
}

func (r *RedisService) GetDeviceFromRedis(ctx context.Context, deviceId string) (*genprotos.GetDeviceResponse, error) {
	deviceJSON, err := r.redisDb.Get(ctx, deviceId).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		r.logger.Printf("ERROR WHILE GETTING DATA FROM REDIS: %s", err.Error())
		return nil, err
	}

	var device genprotos.Device
	err = json.Unmarshal([]byte(deviceJSON), &device)
	if err != nil {
		r.logger.Printf("ERROR WHILE UNMARSHALING DATA: %s", err.Error())
		return nil, err
	}
	return &genprotos.GetDeviceResponse{
		Device: &device,
	}, nil
}

func (r *RedisService) DeleteDeviceFromRedis(ctx context.Context, deviceID string) error {
	result, err := r.redisDb.Del(ctx, deviceID).Result()
	if err != nil {
		r.logger.Printf("ERROR WHILE DELETING DATA FROM REDIS: %s", err.Error())
		return err
	}

	if result == 0 {
		r.logger.Printf("Device with ID %s does not exist in Redis", deviceID)
	} else {
		r.logger.Printf("Device with ID %s has been deleted from Redis", deviceID)
	}

	return nil
}
