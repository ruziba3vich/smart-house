package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	usersprotos "github.com/ruziba3vich/smart-house/genprotos/submodules/users_submodule/protos"
)

func (g *GrpcHandler) GetAllUsers(c *gin.Context) {
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
	c.IndentedJSON(http.StatusOK, gin.H{"response": response})
}
