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

	usersRouter := router.Group("/users")
	usersRouter.POST("/register", a.rbmqHandler.RegisterUser)
	usersRouter.POST("/login", a.rbmqHandler.LoginUser)
	usersRouter.PUT("/:id", middleware.AuthMiddleware(t), a.rbmqHandler.UpdateUser)
	usersRouter.DELETE("/delete/:id", middleware.AuthMiddleware(t), a.rbmqHandler.DeleteUserById)
	usersRouter.GET("/", middleware.AuthMiddleware(t), a.rbmqHandler.GetAllUsers)

	devicesRouter := router.Group("/devices")
	devicesRouter.POST("/", middleware.AuthMiddleware(t), a.rbmqHandler.CreateDevice)
	devicesRouter.PUT("/:id", middleware.AuthMiddleware(t), a.rbmqHandler.UpdateDevice)
	devicesRouter.GET("/:id", middleware.AuthMiddleware(t), a.rbmqHandler.GetDevice)
	devicesRouter.DELETE("/:id", middleware.AuthMiddleware(t), a.rbmqHandler.DeleteDevice)
	devicesRouter.GET("/", middleware.AuthMiddleware(t), a.rbmqHandler.GetAllDevices)

	return router.Run(cfg.Port)
}
