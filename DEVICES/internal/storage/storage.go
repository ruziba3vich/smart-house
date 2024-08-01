package storage

import (
	"context"
	"fmt"

	genprotos "github.com/ruziba3vich/devices/genprotos/devices_submodule"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Storage) CreateDevice(ctx context.Context, req *genprotos.CreateDeviceRequest) (*genprotos.CreateDeviceResponse, error) {
	device := req.Device
	device.Id = primitive.NewObjectID().Hex()

	_, err := s.database.Client.Database("smart_house").Collection("devices").InsertOne(ctx, device)
	if err != nil {
		s.logger.Printf("Failed to insert device: %s", err.Error())
		return nil, err
	}

	return &genprotos.CreateDeviceResponse{Device: device}, nil
}

func (s *Storage) UpdateDevice(ctx context.Context, req *genprotos.UpdateDeviceRequest) (*genprotos.UpdateDeviceResponse, error) {
	device := req.Device
	objectID, err := primitive.ObjectIDFromHex(device.Id)
	if err != nil {
		s.logger.Printf("Failed to convert ID to ObjectID: %s", err.Error())
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": device}

	updateResult, err := s.database.Client.Database("smart_house").Collection("devices").UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Failed to update device: %s", err.Error())
		return nil, err
	}
	if updateResult.ModifiedCount == 0 {
		s.logger.Println("No device was updated")
	}

	return &genprotos.UpdateDeviceResponse{Device: device}, nil
}

func (s *Storage) GetDevice(ctx context.Context, req *genprotos.GetDeviceRequest) (*genprotos.GetDeviceResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		s.logger.Printf("Failed to convert ID to ObjectID: %s", err.Error())
		return nil, err
	}

	var device genprotos.Device
	err = s.database.Client.Database("smart_house").Collection("devices").FindOne(ctx, bson.M{"_id": objectID}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logger.Printf("No device found with ID: %s", req.Id)
			return nil, fmt.Errorf("no device found with ID: %s", req.Id)
		}
		s.logger.Printf("Failed to find device: %s", err.Error())
		return nil, err
	}

	return &genprotos.GetDeviceResponse{Device: &device}, nil
}

func (s *Storage) DeleteDevice(ctx context.Context, req *genprotos.DeleteDeviceRequest) (*genprotos.DeleteDeviceResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		s.logger.Printf("Failed to convert ID to ObjectID: %s", err.Error())
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"deleted": true}}

	updateResult, err := s.database.Client.Database("smart_house").Collection("devices").UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Failed to update device: %s", err.Error())
		return nil, err
	}
	if updateResult.ModifiedCount == 0 {
		s.logger.Println("No device was updated")
	}

	return &genprotos.DeleteDeviceResponse{Success: updateResult.ModifiedCount > 0}, nil
}

func (s *Storage) GetAllDevices(ctx context.Context, req *genprotos.GetAllDevicesRequest) (*genprotos.GetAllDevicesResponse, error) {
	skip := (req.Page - 1) * req.Limit

	findOptions := options.Find()
	findOptions.SetLimit(int64(req.Limit))
	findOptions.SetSkip(int64(skip))

	filter := bson.M{"deleted": false}

	cursor, err := s.database.Client.Database("smart_house").Collection("devices").Find(ctx, filter, findOptions)
	if err != nil {
		s.logger.Printf("Failed to find devices: %s", err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var devices []*genprotos.Device
	for cursor.Next(ctx) {
		var device genprotos.Device
		if err := cursor.Decode(&device); err != nil {
			s.logger.Printf("Failed to decode device: %s", err.Error())
			return nil, err
		}
		devices = append(devices, &device)
	}

	if err := cursor.Err(); err != nil {
		s.logger.Printf("Cursor error: %s", err.Error())
		return nil, err
	}

	return &genprotos.GetAllDevicesResponse{Devices: devices}, nil
}
