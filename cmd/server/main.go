package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	answerDelivery "github.com/han/go-ecommerce/internal/answer/delivery"
	answerRepository "github.com/han/go-ecommerce/internal/answer/repository"
	usecaseAnswer "github.com/han/go-ecommerce/internal/answer/usecase"
	"github.com/han/go-ecommerce/internal/database"
	"github.com/han/go-ecommerce/internal/middleware"
	"github.com/han/go-ecommerce/internal/token"

	authDelivery "github.com/han/go-ecommerce/internal/auth/delivery"
	authRepo "github.com/han/go-ecommerce/internal/auth/repository"
	authUsecase "github.com/han/go-ecommerce/internal/auth/usecase"

	questionDelivery "github.com/han/go-ecommerce/internal/question/delivery"
	questionRepo "github.com/han/go-ecommerce/internal/question/repository"
	questionUsecase "github.com/han/go-ecommerce/internal/question/usecase"
)

func main() {
	// Init DB
	db, err := database.InitDB("cmd/data/gorm.db")
	if err != nil {
		log.Fatalf("cannot initialize database: %v", err)
	}

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment")
	}
	jwtSecret := os.Getenv("JWT_SECRET")

	// Init JWT service
	jwtService, err := token.New(jwtSecret, 12*time.Hour)
	if err != nil {
		log.Fatalf("cannot initialize JWT service: %v", err)
	}

	// Init Gin router
	r := gin.Default()

	// ===== Auth Module =====
	authRepository := authRepo.NewRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)
	authService := authUsecase.New(authRepository, refreshTokenRepository, jwtService, 7*24*time.Hour)
	authHandler := authDelivery.New(authService)
	authDelivery.RegisterRoutes(r, authHandler, middleware.RequireAuth(jwtService))

	// ===== Question Module =====
	questionRepository := questionRepo.New(db)
	questionService := questionUsecase.New(questionRepository)
	questionHandler := questionDelivery.New(questionService)
	questionDelivery.RegisterRoutes(r, questionHandler, middleware.RequireAuth(jwtService))

	// ===== Answer Module =====
	// Repository cho Answer
	// Usecase cho Question (logic nghiệp vụ)
	// Handler cho Question
	// Đăng ký route cho Question, kèm middleware RequireAuth
	answerRepository := answerRepository.New(db)
	answerUc := usecaseAnswer.New(answerRepository)
	answerHandler := answerDelivery.New(answerUc)
	answerDelivery.RegisterRoutes(r, answerHandler, middleware.RequireAuth(jwtService))

	// Start server
	log.Println("Server started at :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
