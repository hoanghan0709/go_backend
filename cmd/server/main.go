package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	authDelivery "github.com/han/go-ecommerce/internal/auth/delivery"
	authRepo "github.com/han/go-ecommerce/internal/auth/repository"
	authSer "github.com/han/go-ecommerce/internal/auth/usecase"
	"github.com/han/go-ecommerce/internal/database"
	"github.com/han/go-ecommerce/internal/middleware"
	questionDelivery "github.com/han/go-ecommerce/internal/question/delivery"
	questionRepo "github.com/han/go-ecommerce/internal/question/repository"
	questionUsecase "github.com/han/go-ecommerce/internal/question/usecase"
	"github.com/han/go-ecommerce/internal/token"
	"github.com/joho/godotenv"
	// "github.com/han/go-ecommerce/internal/question/delivery"
	// "github.com/han/go-ecommerce/internal/question/repository"
	// "github.com/han/go-ecommerce/internal/question/usecase"
)

func main() {
	db, err := database.InitDB("cmd/data/gorm.db")
	if err != nil {
		log.Fatalf("cannot initialize databasse: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtService, err := token.New(
		jwtSecret,
		24*time.Hour,
	)
	if err != nil {
		log.Fatalf("cannot initialize JWT service: %v", err)
	}

	r := gin.Default()
	authRepository := authRepo.NewRepository(db)
	authService := authSer.New(authRepository, jwtService)
	authHandler := authDelivery.New(authService)
	authDelivery.RegisterRoutes(r,
		authHandler,
		middleware.RequireAuth(jwtService))

	questionRepository := questionRepo.New(db)
	questionUsecase := questionUsecase.New(questionRepository)
	questionHandler := questionDelivery.New(questionUsecase)
	questionDelivery.RegisterRoutes(r, questionHandler, middleware.RequireAuth(jwtService))

	log.Println("Server started at :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped:s %v", err)
	}
}
