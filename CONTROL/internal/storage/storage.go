package storage

import (
	"context"
	"log"

	controlrpc "ruziba3vich/github.com/control/genprotos/controller_submodule"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	Storage struct {
		database *DB
		logger   *log.Logger
	}
)

func (s *Storage) TurnDeviceOn(ctx context.Context, req *controlrpc.DeviceRequest) (*controlrpc.DeviceResponse, error) {
	collection := s.database.Client.Database("smart_house").Collection("devices")
	filter := bson.M{"_id": req.DeviceId}
	update := bson.M{"$set": bson.M{"status": "on"}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Error turning device on: %v", err)
		return nil, err
	}

	return &controlrpc.DeviceResponse{
		Status:  "success",
		Message: "Device turned on",
	}, nil
}

func (s *Storage) TurnDeviceOff(ctx context.Context, req *controlrpc.DeviceRequest) (*controlrpc.DeviceResponse, error) {
	collection := s.database.Client.Database("smart_house").Collection("devices")
	filter := bson.M{"_id": req.DeviceId}
	update := bson.M{"$set": bson.M{"status": "off"}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Error turning device off: %v", err)
		return nil, err
	}

	return &controlrpc.DeviceResponse{
		Status:  "success",
		Message: "Device turned off",
	}, nil
}

func (s *Storage) AddUserToHouse(ctx context.Context, req *controlrpc.UserRequest) (*controlrpc.HouseResponse, error) {
	collection := s.database.Client.Database("smart_house").Collection("users")
	filter := bson.M{"_id": req.UserId}
	update := bson.M{"$set": bson.M{"houseId": req.HouseId}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Error adding user to house: %v", err)
		return nil, err
	}

	return &controlrpc.HouseResponse{
		Status:  "success",
		Message: "User added to house",
	}, nil
}

func (s *Storage) RemoveUserFromHouse(ctx context.Context, req *controlrpc.UserRequest) (*controlrpc.HouseResponse, error) {
	collection := s.database.Client.Database("smart_house").Collection("users")
	filter := bson.M{"_id": req.UserId}
	update := bson.M{"$unset": bson.M{"houseId": ""}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Error removing user from house: %v", err)
		return nil, err
	}

	return &controlrpc.HouseResponse{
		Status:  "success",
		Message: "User removed from house",
	}, nil
}

func (s *Storage) GetBatteryStatus(ctx context.Context, req *controlrpc.DeviceRequest) (*controlrpc.BatteryResponse, error) {
	collection := s.database.Client.Database("smart_house").Collection("devices")
	filter := bson.M{"_id": req.DeviceId}

	var device struct {
		Battery int `bson:"battery"`
	}

	err := collection.FindOne(ctx, filter).Decode(&device)
	if err != nil {
		s.logger.Printf("Error getting battery status: %v", err)
		return nil, err
	}

	return &controlrpc.BatteryResponse{
		Battery: int32(device.Battery),
	}, nil
}
