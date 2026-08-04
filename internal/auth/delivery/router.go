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
		auth.POST(
			"/register",
			handler.Register,
		)
		auth.POST(
			"/login",
			handler.Login,
		)
	}

	protected := auth.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/getProfile", ///:id
			handler.GetUser)
	}
}
