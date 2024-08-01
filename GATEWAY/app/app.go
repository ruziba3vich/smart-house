package app

import (
	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/smart-house/app/handler"
	"github.com/ruziba3vich/smart-house/internal/config"
	"github.com/ruziba3vich/smart-house/internal/utils"
	middleware "github.com/ruziba3vich/smart-house/midd-ware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type (
	APP struct {
		rbmqHandler *handler.RbmqHandler
	}
)

func New(rbmqHandler *handler.RbmqHandler) *APP {
	return &APP{
		rbmqHandler: rbmqHandler,
	}
}

func (a *APP) RUN(cfg *config.Config, t *utils.TokenGenerator) error {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/users/register", a.rbmqHandler.RegisterUser)
	router.POST("/users/login", a.rbmqHandler.LoginUser)
	router.PUT("/users/:id", middleware.AuthMiddleware(t), a.rbmqHandler.UpdateUser)
	router.DELETE("/users/delete/:id", middleware.AuthMiddleware(t), a.rbmqHandler.DeleteUserById)
	router.GET("/users/", middleware.AuthMiddleware(t), a.rbmqHandler.GetAllUsers)

	return router.Run(cfg.Port)
}
