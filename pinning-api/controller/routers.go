package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/middleware"
)

// NewRouter returns a new router.
func NewRouter(handlers *Handlers) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	v1.Use(middleware.ApiKeyMiddleware(handlers.Config.Auth, handlers.Log, handlers.UsersRepo))
	v1.POST("/pins", handlers.AddPin)
	v1.DELETE("/pins/:requestid", handlers.DeletePinByRequestId)
	v1.GET("/pins/:requestid", handlers.GetPinByRequestId)
	v1.GET("/pins", handlers.GetPins)
	v1.POST("/pins/:requestid", handlers.ReplacePinByRequestId)
	authorized := router.Group("/admin", gin.BasicAuth(gin.Accounts{
		handlers.Config.Auth.AdminUserName: handlers.Config.Auth.AdminPassword,
	}))
	authorized.GET("/peers/", handlers.GetPeers)
	authorized.GET("/alerts/", handlers.GetAlerts)
	return router
}

// Index is the index handler.
func Index(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}
