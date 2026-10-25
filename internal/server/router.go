package server

import (
	"github.com/gin-gonic/gin"
	answerDelivery "github.com/han/go-ecommerce/internal/answer/delivery"
	answerRepo "github.com/han/go-ecommerce/internal/answer/repository"
	answerUsecase "github.com/han/go-ecommerce/internal/answer/usecase"
	authDelivery "github.com/han/go-ecommerce/internal/auth/delivery"
	authRepo "github.com/han/go-ecommerce/internal/auth/repository"
	authUsecase "github.com/han/go-ecommerce/internal/auth/usecase"
	examDelivery "github.com/han/go-ecommerce/internal/exam/delivery"
	examRepo "github.com/han/go-ecommerce/internal/exam/repository"
	examUsecase "github.com/han/go-ecommerce/internal/exam/usecase"
	"github.com/han/go-ecommerce/internal/learning"
	learningDelivery "github.com/han/go-ecommerce/internal/learning/delivery"
	"github.com/han/go-ecommerce/internal/middleware"
	questionDelivery "github.com/han/go-ecommerce/internal/question/delivery"
	questionRepo "github.com/han/go-ecommerce/internal/question/repository"
	questionUsecase "github.com/han/go-ecommerce/internal/question/usecase"
	"github.com/han/go-ecommerce/internal/token"
	"gorm.io/gorm"
	"time"
)

// NewRouter dùng chung cho server thật và integration test, tránh test một bộ
// route khác với bộ route thực sự được deploy.
func NewRouter(db *gorm.DB, jwt *token.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	auth, admin := middleware.RequireAuth(jwt), middleware.RequireAdmin(db)
	authService := authUsecase.New(authRepo.NewRepository(db), authRepo.NewRefreshTokenRepository(db), jwt, 7*24*time.Hour)
	authDelivery.RegisterRoutes(r, authDelivery.New(authService), auth)
	questionDelivery.RegisterRoutes(r, questionDelivery.New(questionUsecase.New(questionRepo.New(db))), auth, admin)
	answerDelivery.RegisterRoutes(r, answerDelivery.New(answerUsecase.New(answerRepo.New(db))), auth, admin)
	exams := examUsecase.New(examRepo.New(db))
	examDelivery.RegisterRoutes(r, examDelivery.New(exams), auth)
	learningDelivery.RegisterRoutes(r, learning.New(db), auth, admin, exams.ExpireUser)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	return r
}
