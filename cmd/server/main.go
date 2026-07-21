package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/database"
	handler "github.com/han/go-ecommerce/internal/module/auth/handler"
	auth "github.com/han/go-ecommerce/internal/module/auth/repository"
	router "github.com/han/go-ecommerce/internal/module/auth/router"
	service "github.com/han/go-ecommerce/internal/module/auth/service"
)

func main() {
	db, err := database.InitDB("gorm.db")
	if err != nil {
		log.Fatalf("cannot initialize database: %v", err)
	}

	r := gin.Default()
	authRepository := auth.NewRepository(db)
	authService := service.NewService(authRepository)
	authHandler := handler.NewHandler(authService)

	router.RegisterRoutes(r, authHandler)

	log.Println("Server started at :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped:s %v", err)
	}
}
