package router

import (
	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/module/auth/handler"
)

func RegisterRoutes(

	r *gin.Engine,

	handler *handler.Handler,
) {

	auth := r.Group("/auth")

	auth.POST(

		"/register",

		handler.Register,
	)
}
