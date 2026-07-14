package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/module/auth/database"
	handler "github.com/han/go-ecommerce/internal/module/auth/handler"
	auth "github.com/han/go-ecommerce/internal/module/auth/repository"
	router "github.com/han/go-ecommerce/internal/module/auth/router"
	service "github.com/han/go-ecommerce/internal/module/auth/service"
)

func main() {
	database := database.InitDB("gorm.db")
	log.Println("User created successfully")

	// khởi tạo Gin
	r := gin.Default()
	// dependency injection
	authRepository := auth.NewRepository(database)
	authService := service.NewService(authRepository)
	authHandler := handler.NewHandler(authService)

	router.RegisterRoutes(r, authHandler)

	log.Println("Server started at :8080")

	r.Run(":8080")

}
