package main

import (
	"log"

	"github.com/gin-gonic/gin"
	db "github.com/han/go-ecommerce/internal/database"

	// models "github.com/han/go-ecommerce/internal/module"
	auth "github.com/han/go-ecommerce/internal/module/auth"
)

func main() {
	database := db.InitDB("gorm.db")

	// err := db.DB.Create(
	// 	&models.User{
	// 		Name:     "Alice1",
	// 		Email:    "alice@example.com",
	// 		Password: "amjl-1234567890",
	// 	},
	// ).Error

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// r := gin.Default()

	// r.GET("/aa", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"status": "ok",
	// 	})
	// })

	// r.Run(":8080")

	log.Println("User created successfully")

	// khởi tạo Gin
	r := gin.Default()
	// dependency injection
	authRepository := auth.NewRepository(database)
	authService := auth.NewService(authRepository)
	authHandler := auth.NewHandler(authService)

	auth.RegisterRoutes(r, authHandler)

	log.Println("Server started at :8080")

	r.Run(":8080")

}
