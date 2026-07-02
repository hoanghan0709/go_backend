package main

import (
	"log"

	"github.com/han/go-ecommerce/internal/database"
	"github.com/han/go-ecommerce/internal/models"
)

func main() {
	database.InitDB("gorm.db")

	err := database.DB.Create(
		&models.User{
			Name:     "Alice1",
			Email:    "alice@example.com",
			Password: "amjl-1234567890",
		},
	).Error

	if err != nil {
		log.Fatal(err)
	}

	// r := gin.Default()

	// r.GET("/aa", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"status": "ok",
	// 	})
	// })

	// r.Run(":8080")

	log.Println("User created successfully")
}
