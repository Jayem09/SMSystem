package main

import (
	"fmt"
	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"smsystem-backend/internal/config"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	fmt.Println("=== DIAGNOSTIC REPORT ===")

	var users []models.User
	database.DB.Preload("Branch").Find(&users)
	fmt.Println("\n[USERS & BRANCHES]")
	for _, u := range users {
		bName := "None (0)"
		if u.BranchID != 0 {
			bName = fmt.Sprintf("Branch ID %d", u.BranchID)
		}
		fmt.Printf("User: %s | Role: %s | Branch: %s\n", u.Name, u.Role, bName)
	}

	var warehouses []models.Warehouse
	database.DB.Preload("Branch").Find(&warehouses)
	fmt.Println("\n[WAREHOUSES]")
	for _, w := range warehouses {
		fmt.Printf("Warehouse: %s | Branch ID: %d\n", w.Name, w.BranchID)
	}

	var batches []models.Batch
	database.DB.Find(&batches)
	fmt.Println("\n[BATCHES]")
	for _, b := range batches {
		fmt.Printf("Batch #%s | Product %d | Qty: %d | Branch ID: %d | Warehouse ID: %d\n", b.BatchNumber, b.ProductID, b.Quantity, b.BranchID, b.WarehouseID)
	}

	var products []models.Product
	database.DB.Find(&products)
	fmt.Println("\n[PRODUCTS (Raw DB Table)]")
	for _, p := range products {
		fmt.Printf("Product %d | %s | DB Stock: %d\n", p.ID, p.Name, p.Stock)
	}
}
