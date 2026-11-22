// Công cụ chạy cục bộ bởi chủ server; không cung cấp API tự nâng quyền.
package main

import (
	"flag"
	"log"
	"os"

	auth "github.com/han/go-ecommerce/internal/auth/model"
	"github.com/han/go-ecommerce/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	email := flag.String("email", "", "existing account email")
	role := flag.String("role", "admin", "admin or learner")
	flag.Parse()
	if *email == "" || (*role != "admin" && *role != "learner") {
		log.Fatal("provide -email and -role admin|learner")
	}
	_ = godotenv.Load()
	db, err := database.Open(os.Getenv("DATABASE_URL"), os.Getenv("DATABASE_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	var user auth.User
	if err = db.Where("email = ?", *email).First(&user).Error; err != nil {
		log.Fatal(err)
	}
	if err = db.Model(&user).Update("role", *role).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("Updated role to %s for user %d", *role, user.ID)
}
