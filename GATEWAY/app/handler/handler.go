package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
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
		logger      *log.Logger
		Msgbroker   *msgbroker.MsgBroker
		tokenizer   *utils.TokenGenerator
		usersClient usersprotos.UsersServiceClient
		cfg         *config.Config
		rq          amqp.Queue
		uq          amqp.Queue
		dq          amqp.Queue
	}
)

func NewRbmqHandler(logger *log.Logger,
	msgbroker *msgbroker.MsgBroker,
	tokenizer *utils.TokenGenerator,
	usersClient usersprotos.UsersServiceClient,
	cfg *config.Config,
	rq amqp.Queue,
	uq amqp.Queue,
	dq amqp.Queue) *RbmqHandler {
	return &RbmqHandler{
		logger:      logger,
		Msgbroker:   msgbroker,
		usersClient: usersClient,
		tokenizer:   tokenizer,
		cfg:         cfg,
		rq:          rq,
		uq:          uq,
		dq:          dq,
	}
}

// RegisterUser godoc
// @Summary Register
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.User true "User registration information"
// @Security ApiKeyAuth
// @Success 201 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /users/register [post]
func (r *RbmqHandler) RegisterUser(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Println("ERROR WHILE MARSHALING DATA")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}

	err = r.Msgbroker.PublishToQueue(r.Msgbroker.Registration, body, r.rq, "create_reply", r.cfg.ContentType)
	if err != nil {
		r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	user, err := r.usersClient.GetByEmail(c, &usersprotos.GetByFieldRequest{
		GetByField: req.Email,
	})
	if err != nil {
		r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"response": user})
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
// @Success 201 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /users/{id} [put]
func (r *RbmqHandler) UpdateUser(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Println("ERROR WHILE BINDING DATA")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	strUserId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		r.logger.Println("ERROR WHILE GETTING USER ID")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	req.Id = strUserId
	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Println("ERROR WHILE MARSHALING DATA")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}

	r.Msgbroker.PublishToQueue(r.Msgbroker.Updates, body, r.uq, "update_reply", r.cfg.ContentType)
	// if err != nil {
	// 	r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
	// 	c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
	// 	return
	// }
	user, err := r.usersClient.GetById(c, &usersprotos.GetByFieldRequest{
		GetByField: req.Id.Hex(),
	})
	if err != nil {
		r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"response": user})
}

// DeleteUserById godoc
// @Summary Delete
// @Description delete an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security ApiKeyAuth
// @Success 201 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /users/delete/{id} [delete]
func (r *RbmqHandler) DeleteUserById(c *gin.Context) {
	req := models.DeleteUserRequest{
		UserId: c.Param("id"),
	}

	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Println("ERROR WHILE MARSHALING DATA")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}

	r.Msgbroker.PublishToQueue(r.Msgbroker.Deletes, body, r.dq, "delete_reply", r.cfg.ContentType)
	// if err != nil {
	// 	r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
	// 	c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
	// 	return
	// }
	user, err := r.usersClient.GetById(c, &usersprotos.GetByFieldRequest{
		GetByField: req.UserId,
	})
	if err != nil {
		r.logger.Println("ERROR HAS BEEN RETURNED FROM THE SERVER", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"response": user})
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Retrieve all users
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H{"response": usersprotos.GetAllUsersResponse}
// @Failure 500 {object} gin.H{"error": string}
// @Router /users [get]
func (g *RbmqHandler) GetAllUsers(c *gin.Context) {
	var req usersprotos.GetAllUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.Println("ERROR WHILE BINDING DATA :", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	response, err := g.usersClient.GetAllUsers(c, &req)
	if err != nil {
		g.logger.Println("ERROR RETURNED FROM THE SERVER :", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.IndentedJSON(http.StatusOK, response.Users)
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
