package storage

import (
	"context"
	"fmt"

	genprotos "github.com/ruziba3vich/users/genprotos/users_submodule/protos"
	"github.com/ruziba3vich/users/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateUser inserts a new user into the collection
func (s *Storage) CreateUser(ctx context.Context, req *genprotos.CreateUserReuest) (*genprotos.User, error) {
	var user models.User
	user.Id = primitive.NewObjectID()
	user.FromProto(req)
	user.Password = s.passwordHasher.HashPassword(user.Password)
	s.logger.Println(user.Password, "-------------------------------------")

	select {
	case <-ctx.Done():
		s.logger.Println("context cancelled or expired")
		return nil, ctx.Err()
	default:
	}

	_, err := s.database.UsersCollection.InsertOne(ctx, user)
	if err != nil {
		s.logger.Printf("Failed to insert user: %s\n", err.Error())
		return nil, fmt.Errorf("failed to insert user: %s", err.Error())
	}
	s.logger.Printf("--------------------- USER HAS BEEN CREATED WITH EMAil %s -----------------------\n", user.Email)
	return user.ToProtoUser(), nil
}

// UpdateUser updates an existing user
func (s *Storage) UpdateUser(ctx context.Context, req *genprotos.UpdateUserReuqest) (*genprotos.User, error) {
	if req.User == nil || req.User.UserId == "" {
		return nil, fmt.Errorf("invalid update request: user or user ID is missing")
	}

	user, err := s.getByField(ctx, &models.GetByFieldRequest{
		Field: "_id",
		Value: req.User.UserId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %s", err.Error())
	}

	user.Update(req)
	objectId, err := primitive.ObjectIDFromHex(req.User.UserId)
	if err != nil {
		s.logger.Println("failed to convert the object ID:", err.Error())
		return nil, fmt.Errorf("invalid ObjectID: %s", err.Error())
	}
	filter := bson.M{"_id": objectId, "deleted": false}

	update := bson.M{
		"$set": user,
	}

	updateResult, err := s.database.UsersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Failed to update document: %s", err.Error())
		return nil, fmt.Errorf("failed to update user: %s", err.Error())
	}
	if updateResult.ModifiedCount == 0 {
		s.logger.Println("no rows updated")
		return nil, fmt.Errorf("no user found to update with ID: %s", req.User.UserId)
	}
	return user.ToProtoUser(), nil
}

// getByField finds a user by a specific field
func (s *Storage) getByField(ctx context.Context, req *models.GetByFieldRequest) (*models.User, error) {
	var user models.User
	var filter bson.M

	if req.Field == "_id" {
		objectID, err := primitive.ObjectIDFromHex(req.Value)
		if err != nil {
			s.logger.Printf("Invalid ObjectID: %s\n", req.Value)
			return nil, fmt.Errorf("invalid ObjectID: %s", req.Value)
		}
		filter = bson.M{"_id": objectID, "deleted": false}
	} else {
		filter = bson.M{req.Field: req.Value, "deleted": false}
	}

	select {
	case <-ctx.Done():
		s.logger.Println("context canceled or expired")
		return nil, ctx.Err()
	default:
	}

	err := s.database.UsersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logger.Printf("No document found with %s: %s\n", req.Field, req.Value)
			return nil, fmt.Errorf("no document found with %s: %s", req.Field, req.Value)
		}
		s.logger.Printf("Failed to find document: %s", err.Error())
		return nil, fmt.Errorf("failed to find document: %s", err.Error())
	}

	return &user, nil
}

// GetUserByUsername gets a user by username
func (s *Storage) GetUserByUsername(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.User, error) {
	request := models.GetByFieldRequest{
		Field: "username",
		Value: req.GetByField,
	}
	user, err := s.getByField(ctx, &request)
	if err != nil {
		return nil, err
	}
	return user.ToProtoUser(), nil
}

// GetUserById gets a user by ID
func (s *Storage) GetUserById(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.User, error) {
	request := models.GetByFieldRequest{
		Field: "_id",
		Value: req.GetByField,
	}
	user, err := s.getByField(ctx, &request)
	if err != nil {
		return nil, err
	}
	return user.ToProtoUser(), nil
}

// GetUserByEmail gets a user by email
func (s *Storage) GetUserByEmail(ctx context.Context, req *genprotos.GetByFieldRequest) (*genprotos.User, error) {
	request := models.GetByFieldRequest{
		Field: "email",
		Value: req.GetByField,
	}
	user, err := s.getByField(ctx, &request)
	if err != nil {
		return nil, err
	}
	return user.ToProtoUser(), nil
}

// GetUserByAddress gets users by a specific address
func (s *Storage) GetUserByAddress(ctx context.Context, req *genprotos.GetUsersByAddressRequest) (*genprotos.GetAllUsersResponse, error) {
	filter := bson.M{"profile.address": req.Address, "deleted": false}

	cursor, err := s.database.UsersCollection.Find(ctx, filter)
	if err != nil {
		s.logger.Printf("Failed to find users: %s", err.Error())
		return nil, fmt.Errorf("failed to find users: %s", err.Error())
	}
	defer cursor.Close(ctx)

	var response genprotos.GetAllUsersResponse

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			s.logger.Printf("Failed to decode user: %s", err.Error())
			return nil, fmt.Errorf("failed to decode user: %s", err.Error())
		}
		response.Users = append(response.Users, user.ToProtoUser())
	}

	if err := cursor.Err(); err != nil {
		s.logger.Printf("Cursor error: %s", err.Error())
		return nil, fmt.Errorf("cursor error: %s", err.Error())
	}

	return &response, nil
}

// GetAllUsers gets all users with pagination
func (s *Storage) GetAllUsers(ctx context.Context, req *genprotos.GetAllUsersRequest) (*genprotos.GetAllUsersResponse, error) {
	skip := (req.Pagination - 1) * req.Limit

	findOptions := options.Find()
	findOptions.SetLimit(int64(req.Limit))
	findOptions.SetSkip(int64(skip))

	filter := bson.M{"deleted": false}

	cursor, err := s.database.UsersCollection.Find(ctx, filter, findOptions)
	if err != nil {
		s.logger.Printf("FAILED TO FIND USERS: %s", err.Error())
		return nil, fmt.Errorf("failed to find users: %s", err.Error())
	}
	defer cursor.Close(ctx)

	var response genprotos.GetAllUsersResponse
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			s.logger.Printf("Failed to decode user: %s", err.Error())
			return nil, fmt.Errorf("failed to decode user: %s", err.Error())
		}
		response.Users = append(response.Users, user.ToProtoUser())
	}

	if err := cursor.Err(); err != nil {
		s.logger.Printf("Cursor error: %s", err.Error())
		return nil, fmt.Errorf("cursor error: %s", err.Error())
	}

	return &response, nil
}

// DeleteUserById marks a user as deleted by ID
func (s *Storage) DeleteUserById(ctx context.Context, req *genprotos.GetByFieldRequest) error {
	objectID, err := primitive.ObjectIDFromHex(req.GetByField)
	if err != nil {
		s.logger.Printf("Invalid ObjectID: %s\n", req.GetByField)
		return fmt.Errorf("invalid ObjectID: %s", req.GetByField)
	}

	filter := bson.M{"_id": objectID}

	var user models.User
	err = s.database.UsersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logger.Printf("No document found with ID: %s\n", req.GetByField)
			return fmt.Errorf("no document found with ID: %s", req.GetByField)
		}
		s.logger.Printf("Failed to find user: %s", err.Error())
		return fmt.Errorf("failed to find user: %s", err.Error())
	}

	update := bson.M{
		"deleted": true,
	}
	_, err = s.database.UsersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Printf("Failed to update user: %s", err.Error())
		return fmt.Errorf("failed to update user: %s", err.Error())
	}

	user.Deleted = true

	s.logger.Printf("Successfully marked user as deleted with ID: %s\n", req.GetByField)
	return nil
}

// LoginUser authenticates a user and returns a token
func (s *Storage) LoginUser(ctx context.Context, req *genprotos.LoginRequest) (*genprotos.RegisterUserResponse, error) {
	user, err := s.GetUserByEmail(ctx, &genprotos.GetByFieldRequest{
		GetByField: req.Email,
	})
	if err != nil {
		return nil, err
	}
	s.logger.Println("----------------------- user found -----------------------------")
	if s.passwordHasher.CheckPasswordHash(req.Password, user.Password) {
		token, err := s.tokenGenerator.GenerateToken(user.UserId, user.Username)
		if err != nil {
			s.logger.Printf("ERROR WHILE GENERATING TOKEN FOR USER %s\n", user.Email)
			return nil, err
		}
		return &genprotos.RegisterUserResponse{
			User: user,
			Token: &genprotos.Token{
				StringToken: token,
			},
		}, nil
	}
	return nil, fmt.Errorf("mismatch in password")
}
