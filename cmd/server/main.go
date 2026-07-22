package main

import (
	"log"

	"github.com/gin-gonic/gin"
	authDelivery "github.com/han/go-ecommerce/internal/auth/delivery"
	authRepo "github.com/han/go-ecommerce/internal/auth/repository"
	authSer "github.com/han/go-ecommerce/internal/auth/usecase"
	"github.com/han/go-ecommerce/internal/database"
	// "github.com/han/go-ecommerce/internal/question/delivery"
	// "github.com/han/go-ecommerce/internal/question/repository"
	// "github.com/han/go-ecommerce/internal/question/usecase"
)

func main() {
	db, err := database.InitDB("cmd/data/gorm.db")
	if err != nil {
		log.Fatalf("cannot initialize databasse: %v", err)
	}

	r := gin.Default()
	authRepository := authRepo.NewRepository(db)
	authService := authSer.New(authRepository)
	authHandler := authDelivery.New(authService)
	authDelivery.RegisterRoutes(r, authHandler)

	// questionRepository := repository.New(db)
	// questionUsecase := usecase.New(questionRepository)
	// questionHandler := delivery.New(questionUsecase)
	// delivery.RegisterRoutes(r, questionHandler)

	log.Println("Server started at :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped:s %v", err)
	}
}
