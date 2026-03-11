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

	var branches []models.Branch
	database.DB.Find(&branches)
	fmt.Printf("Total Branches: %d\n", len(branches))
	for _, b := range branches {
		fmt.Printf("- [%d] %s (%s) Active: %v\n", b.ID, b.Name, b.Code, b.IsActive)
	}

	var users []models.User
	database.DB.Preload("Branch").Find(&users)
	fmt.Printf("\nTotal Users: %d\n", len(users))
	for _, u := range users {
		branchName := "None"
		if u.Branch.ID != 0 {
			branchName = u.Branch.Name
		}
		fmt.Printf("- %s (%s) Role: %s Branch: %s (ID: %d)\n", u.Name, u.Email, u.Role, branchName, u.BranchID)
	}
}
