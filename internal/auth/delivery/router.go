package delivery

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(

	r *gin.Engine,

	handler *Handler,
) {

	auth := r.Group("/auth")

	auth.POST(
		"/register",
		handler.Register,
	)
	auth.GET("/getProfile/:id",
		handler.GetUser)
}
