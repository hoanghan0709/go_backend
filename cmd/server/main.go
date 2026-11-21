package main

import (
	"log"
	"os"
	"time"

	"github.com/han/go-ecommerce/internal/database"
	"github.com/han/go-ecommerce/internal/server"
	"github.com/han/go-ecommerce/internal/token"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment")
	}
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "cmd/data/gorm.db"
	}
	db, err := database.InitDB(path)
	if err != nil {
		log.Fatal(err)
	}
	jwt, err := token.New(os.Getenv("JWT_SECRET"), 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = ":8080"
	}
	if err = server.NewRouter(db, jwt).Run(address); err != nil {
		log.Fatal(err)
	}
}
