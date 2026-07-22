package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/auth/delivery"
	"github.com/han/go-ecommerce/internal/auth/repository"
	"github.com/han/go-ecommerce/internal/auth/usecase"
	"github.com/han/go-ecommerce/internal/database"
)

func main() {
	db, err := database.InitDB("gorm.db")
	if err != nil {
		log.Fatalf("cannot initialize databasse: %v", err)
	}

	r := gin.Default()
	authRepository := repository.NewRepository(db)
	authService := usecase.NewService(authRepository)
	authHandler := delivery.NewHandler(authService)

	delivery.RegisterRoutes(r, authHandler)

	log.Println("Server started at :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped:s %v", err)
	}
}
