package app

import (
	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/smart-house/app/handler"
	"github.com/ruziba3vich/smart-house/internal/config"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type (
	APP struct {
		rbmqHandler *handler.RbmqHandler
		grpcHandler *handler.GrpcHandler
	}
)

func New(rbmqHandler *handler.RbmqHandler, grpcHandler *handler.GrpcHandler) *APP {
	return &APP{
		rbmqHandler: rbmqHandler,
		grpcHandler: grpcHandler,
	}
}

func (a *APP) RUN(cfg *config.Config) error {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/users/register", a.rbmqHandler.RegisterUser)
	router.POST("/users/update/:id", a.rbmqHandler.UpdateUser)
	router.DELETE("/users/delete/:id", a.rbmqHandler.DeleteUserById)
	router.GET("/users/", a.grpcHandler.GetAllUsers)

	return router.Run(cfg.Port)
}
