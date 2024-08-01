package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
	amqp "github.com/rabbitmq/amqp091-go"
	controlrpc "github.com/ruziba3vich/smart-house/genprotos/controller_submodule"
	devicesrpc "github.com/ruziba3vich/smart-house/genprotos/devices_submodule"
	usersprotos "github.com/ruziba3vich/smart-house/genprotos/submodules/users_submodule/protos"

	"github.com/ruziba3vich/smart-house/internal/config"
	models "github.com/ruziba3vich/smart-house/internal/modules"
	"github.com/ruziba3vich/smart-house/internal/msgbroker"
	"github.com/ruziba3vich/smart-house/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"

	_ "github.com/ruziba3vich/smart-house/docs"
)

type (
	RbmqHandler struct {
		logger           *log.Logger
		Msgbroker        *msgbroker.MsgBroker
		tokenizer        *utils.TokenGenerator
		usersClient      usersprotos.UsersServiceClient
		devicesClient    devicesrpc.DeviceServiceClient
		controllerClient controlrpc.ControllerServiceClient
		cfg              *config.Config
		rq               amqp.Queue
		uq               amqp.Queue
		dq               amqp.Queue
	}
)

func NewRbmqHandler(logger *log.Logger,
	msgbroker *msgbroker.MsgBroker,
	tokenizer *utils.TokenGenerator,
	usersClient usersprotos.UsersServiceClient,
	devicesClient devicesrpc.DeviceServiceClient,
	controllerClient controlrpc.ControllerServiceClient,
	cfg *config.Config,
	rq amqp.Queue,
	uq amqp.Queue,
	dq amqp.Queue) *RbmqHandler {
	return &RbmqHandler{
		logger:           logger,
		Msgbroker:        msgbroker,
		usersClient:      usersClient,
		devicesClient:    devicesClient,
		controllerClient: controllerClient,
		tokenizer:        tokenizer,
		cfg:              cfg,
		rq:               rq,
		uq:               uq,
		dq:               dq,
	}
}

// @title Artisan Connect
// @version 1.0
// @description This is a sample server for a restaurant reservation system.
// @securityDefinitions.apikey Bearer
// @in         header
// @name Authorization
// @description Enter the token in the format `Bearer {token}`
// @host localhost:7777
// @BasePath /

// LoginUser godoc
// @Summary Login
// @Description Login an existing user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body usersprotos.LoginRequest true "User login information"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/login [post]
func (r *RbmqHandler) LoginUser(c *gin.Context) {
	var req usersprotos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}
	response, err := r.usersClient.LoginUser(c, &req)
	pp.Println(&response)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.UserResponse{Response: response})
}

// RegisterUser godoc
// @Summary Register
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.User true "User registration information"
// @Security ApiKeyAuth
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/register [post]
func (r *RbmqHandler) RegisterUser(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	exists, err := r.checkIfUserExists(c, &req)
	if exists {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: err.Error()})
		return
	}

	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Println("ERROR WHILE MARSHALING DATA: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	err = r.Msgbroker.PublishToQueue(r.Msgbroker.Registration, body, r.rq, "create_reply", r.cfg.ContentType)
	if err != nil {
		r.logger.Println("-- ERROR FROM SERVER -- `: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	select {
	case <-time.After(time.Second * 5):
		user, err := r.usersClient.GetByEmail(c, &usersprotos.GetByFieldRequest{GetByField: req.Email})
		if err != nil {
			r.logger.Println("ERROR FROM SERVER: ", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusCreated, models.UserResponse{Response: user})
	case <-time.After(time.Second * 10):
		r.logger.Println("ERROR: TIMEOUT WAITING FOR USER CREATION")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Timeout waiting for user creation"})
		return
	}
}

// UpdateUser godoc
// @Summary Update
// @Description Update existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body models.User true "User update information"
// @Security ApiKeyAuth
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
func (r *RbmqHandler) UpdateUser(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	strUserId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		r.logger.Println("ERROR WHILE GETTING USER ID")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	req.Id = strUserId
	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Println("ERROR WHILE MARSHALING DATA")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	r.Msgbroker.PublishToQueue(r.Msgbroker.Updates, body, r.uq, "update_reply", r.cfg.ContentType)
	user, err := r.usersClient.GetById(c, &usersprotos.GetByFieldRequest{
		GetByField: req.Id.Hex(),
	})
	if err != nil {
		r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.UserResponse{Response: user})
}

// DeleteUserById godoc
// @Summary Delete
// @Description delete an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security ApiKeyAuth
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/delete/{id} [delete]
func (r *RbmqHandler) DeleteUserById(c *gin.Context) {
	req := models.DeleteUserRequest{
		UserId: c.Param("id"),
	}

	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Println("ERROR WHILE MARSHALING DATA")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	r.Msgbroker.PublishToQueue(r.Msgbroker.Deletes, body, r.dq, "delete_reply", r.cfg.ContentType)
	user, err := r.usersClient.GetById(c, &usersprotos.GetByFieldRequest{
		GetByField: req.UserId,
	})
	if err != nil {
		r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.UserResponse{Response: user})
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Retrieve all users
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} []models.User
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func (g *RbmqHandler) GetAllUsers(c *gin.Context) {
	var req usersprotos.GetAllUsersRequest

	resp, err := g.usersClient.GetAllUsers(c, &req)
	if err != nil {
		g.logger.Println("ERROR RETURNED FROM THE SERVER :", err.Error())
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	var response []*models.User
	for i := range resp.Users {
		var user models.User
		user.FromProtoUser(resp.Users[i])
		response = append(response, &user)
	}
	c.JSON(http.StatusOK, response)
}

/*
   rpc RegisterUser(CreateUserReuest) returns (Response); ///
   rpc LoginUser(LoginRequest) returns (RegisterUserResponse);
   rpc GetById(GetByFieldRequest) returns (User);
   rpc GetByUsername(GetByFieldRequest) returns (User);
   rpc GetByEmail(GetByFieldRequest) returns (User);
   rpc UpdateUser(UpdateUserReuqest) returns (Response); ///
   rpc GetAllUsers(google.protobuf.Empty) returns (GetAllUsersResponse);
   rpc DeleteUserById(GetByFieldRequest) returns (Response); ///
   rpc GetUsersByAddress(GetUsersByAddressRequest) returns (GetAllUsersResponse);
*/

func (r *RbmqHandler) checkIfUserExists(ctx context.Context, req *models.User) (bool, error) {
	someReq := usersprotos.GetByFieldRequest{GetByField: req.Email}
	user, _ := r.usersClient.GetByEmail(ctx, &someReq)
	if user != nil {
		r.logger.Printf("USER WITH EMAIL %s ALREADY EXISTS\n", req.Email)
		return true, fmt.Errorf("user with email %s already exists", req.Email)
	}

	someReq.GetByField = req.Username
	user, _ = r.usersClient.GetByUsername(ctx, &someReq)
	if user != nil {
		r.logger.Printf("USER WITH USERNAME %s ALREADY EXISTS\n", req.Username)
		return true, fmt.Errorf("user with username %s already exists", req.Username)
	}
	return false, nil
}

// CreateDevice godoc
// @Summary Create a new device
// @Description Create a new device
// @Tags devices
// @Accept json
// @Produce json
// @Param body body devicesprotos.CreateDeviceRequest true "Device creation information"
// @Success 201 {object} devicesprotos.CreateDeviceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /devices [post]
func (r *RbmqHandler) CreateDevice(c *gin.Context) {
	var req devicesrpc.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}
	response, err := r.devicesClient.CreateDevice(c, &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateDevice godoc
// @Summary Update an existing device
// @Description Update an existing device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param body body devicesprotos.UpdateDeviceRequest true "Device update information"
// @Success 200 {object} devicesprotos.UpdateDeviceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /devices/{id} [put]
func (r *RbmqHandler) UpdateDevice(c *gin.Context) {
	var req devicesrpc.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}
	req.Device.Id = c.Param("id")
	response, err := r.devicesClient.UpdateDevice(c, &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetDevice godoc
// @Summary Get a device by ID
// @Description Retrieve a device by ID
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} devicesprotos.GetDeviceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /devices/{id} [get]
func (r *RbmqHandler) GetDevice(c *gin.Context) {
	req := devicesrpc.GetDeviceRequest{Id: c.Param("id")}
	response, err := r.devicesClient.GetDevice(c, &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// DeleteDevice godoc
// @Summary Delete a device by ID
// @Description Delete a device by ID
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} devicesprotos.DeleteDeviceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /devices/{id} [delete]
func (r *RbmqHandler) DeleteDevice(c *gin.Context) {
	req := devicesrpc.DeleteDeviceRequest{Id: c.Param("id")}
	response, err := r.devicesClient.DeleteDevice(c, &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetAllDevices godoc
// @Summary Get all devices
// @Description Retrieve all devices with pagination
// @Tags devices
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Items per page"
// @Success 200 {object} devicesprotos.GetAllDevicesResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /devices [get]
func (r *RbmqHandler) GetAllDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	req := devicesrpc.GetAllDevicesRequest{Page: int32(page), Limit: int32(limit)}
	response, err := r.devicesClient.GetAllDevices(c, &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Turn on a device
// @Description Turn on a device by ID
// @Accept json
// @Produce json
// @Param request body controlrpc.DeviceRequest true "Device Request"
// @Success 200 {object} controlrpc.DeviceResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /devices/on [post]
func (r *RbmqHandler) TurnDeviceOn(c *gin.Context) {
	var req controlrpc.DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := r.controllerClient.TurnDeviceOn(context.Background(), &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Turn off a device
// @Description Turn off a device by ID
// @Accept json
// @Produce json
// @Param request body controlrpc.DeviceRequest true "Device Request"
// @Success 200 {object} controlrpc.DeviceResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /devices/off [post]
func (r *RbmqHandler) TurnDeviceOff(c *gin.Context) {
	var req controlrpc.DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := r.controllerClient.TurnDeviceOff(context.Background(), &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Add user to house
// @Description Add a user to a house by IDs
// @Accept json
// @Produce json
// @Param request body controlrpc.UserRequest true "User Request"
// @Success 200 {object} controlrpc.HouseResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/add [post]
func (r *RbmqHandler) AddUserToHouse(c *gin.Context) {
	var req controlrpc.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := r.controllerClient.AddUserToHouse(context.Background(), &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Remove user from house
// @Description Remove a user from a house by IDs
// @Accept json
// @Produce json
// @Param request body controlrpc.UserRequest true "User Request"
// @Success 200 {object} controlrpc.HouseResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/remove [post]
func (r *RbmqHandler) RemoveUserFromHouse(c *gin.Context) {
	var req controlrpc.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := r.controllerClient.RemoveUserFromHouse(context.Background(), &req)
	if err != nil {
		r.logger.Println("ERROR FROM SERVER: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
