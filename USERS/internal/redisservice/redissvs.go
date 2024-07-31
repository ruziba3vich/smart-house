package redisservice

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	genprotos "github.com/ruziba3vich/users/genprotos/users_submodule/protos"
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

func (r *RedisService) StoreUserInRedis(ctx context.Context, user *genprotos.User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = r.redisDb.Set(ctx, user.UserId, userJSON, time.Hour*24).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisService) GetUserFromRedis(ctx context.Context, userId string) (*genprotos.User, error) {
	userJSON, err := r.redisDb.Get(ctx, userId).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		r.logger.Printf("ERROR WHILE GETTING DATA FROM REDIS : %s\n", err.Error())
		return nil, err
	}

	var user genprotos.User
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		r.logger.Printf("ERROR WHILE MARSHALING DATA : %s\n", err.Error())
		return nil, err
	}
	return &user, nil
}

func (r *RedisService) DeleteUserFromRedis(ctx context.Context, userID string) error {
	result, err := r.redisDb.Del(ctx, userID).Result()
	if err != nil {
		return err
	}

	if result == 0 {
		r.logger.Printf("User with ID %s does not exist in Redis", userID)
	} else {
		r.logger.Printf("User with ID %s has been deleted from Redis", userID)
	}

	return nil
}
