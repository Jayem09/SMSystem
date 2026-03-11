package main

import (
	"fmt"
	"log"
	"strings"

	"smsystem-backend/internal/config"
	"smsystem-backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	dsn := cfg.DatabaseURL
	if strings.HasPrefix(dsn, "mysql://") {
		dsn = strings.TrimPrefix(dsn, "mysql://")
		parts := strings.SplitN(dsn, "@", 2)
		if len(parts) == 2 {
			dsn = fmt.Sprintf("%s@tcp(%s)/%s", parts[0], strings.SplitN(parts[1], "/", 2)[0], strings.SplitN(parts[1], "/", 2)[1])
		}
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var branches []models.Branch
	db.Find(&branches)
	fmt.Printf("Total Branches: %d\n", len(branches))
	for _, b := range branches {
		fmt.Printf("- [%d] %s (%s)\n", b.ID, b.Name, b.Code)
	}

	var users []models.User
	db.Find(&users)
	fmt.Printf("\nTotal Users: %d\n", len(users))
	for _, u := range users {
		fmt.Printf("- %s (%s) Role: %s BranchID: %d\n", u.Name, u.Email, u.Role, u.BranchID)
	}
}
