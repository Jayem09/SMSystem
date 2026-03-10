package main

import (
	"fmt"
	"smsystem-backend/internal/config"
	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)

	fmt.Println("--- DATABASE DIAGNOSTIC ---")

	var branches []models.Branch
	database.DB.Find(&branches)
	fmt.Printf("Branches found: %d\n", len(branches))
	for _, b := range branches {
		fmt.Printf("  - Branch [%d]: %s (%s)\n", b.ID, b.Name, b.Code)
	}

	var users []models.User
	database.DB.Find(&users)
	fmt.Printf("Users found: %d\n", len(users))
	for _, u := range users {
		fmt.Printf("  - User [%d]: %s (%s) Role: %s BranchID: %d\n", u.ID, u.Name, u.Email, u.Role, u.BranchID)
	}

	fmt.Println("---------------------------")
}
