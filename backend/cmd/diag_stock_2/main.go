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

	fmt.Println("=== Latest 3 Products ===")
	var products []models.Product
	database.DB.Order("id desc").Limit(3).Find(&products)
	for _, p := range products {
		fmt.Printf("ID: %d, Name: %s, Table-Stock: %d\n", p.ID, p.Name, p.Stock)
	}

	fmt.Println("\n=== Latest 3 Batches ===")
	var batches []models.Batch
	database.DB.Order("id desc").Limit(3).Find(&batches)
	for _, b := range batches {
		fmt.Printf("ID: %d, ProdID: %d, BranchID: %d, Qty: %d, Num: %s\n", b.ID, b.ProductID, b.BranchID, b.Quantity, b.BatchNumber)
	}

	fmt.Println("\n=== Latest 3 Stock Movements ===")
	var movements []models.StockMovement
	database.DB.Order("id desc").Limit(3).Find(&movements)
	for _, m := range movements {
		fmt.Printf("ID: %d, ProdID: %d, BranchID: %d, Qty: %d, Type: %s, Ref: %s\n", m.ID, m.ProductID, m.BranchID, m.Quantity, m.Type, m.Reference)
	}

	fmt.Println("\n=== Warehouses ===")
	var warehouses []models.Warehouse
	database.DB.Find(&warehouses)
	for _, w := range warehouses {
		fmt.Printf("ID: %d, Name: %s, BranchID: %d\n", w.ID, w.Name, w.BranchID)
	}
}
