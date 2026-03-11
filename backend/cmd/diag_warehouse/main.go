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

	fmt.Println("=== Warehouses ===")
	var warehouses []models.Warehouse
	database.DB.Find(&warehouses)
	for _, w := range warehouses {
		fmt.Printf("ID: %d, Name: %s, BranchID: %d\n", w.ID, w.Name, w.BranchID)
	}

	fmt.Println("\n=== Branches ===")
	var branches []models.Branch
	database.DB.Find(&branches)
	for _, b := range branches {
		fmt.Printf("ID: %d, Name: %s\n", b.ID, b.Name)
	}
}
