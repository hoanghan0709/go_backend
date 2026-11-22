package delivery

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	handler *Handler,
	authMiddleware gin.HandlerFunc,
) {

	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
	}

	protected := auth.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/getProfile", handler.GetUser) //:id
	}

}
