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

	fmt.Println("=== Products ===")
	var products []models.Product
	database.DB.Order("id desc").Limit(5).Find(&products)
	for _, p := range products {
		fmt.Printf("ID: %d, Name: %s, Table-Stock: %d\n", p.ID, p.Name, p.Stock)
	}

	fmt.Println("\n=== Batches (Latest 10) ===")
	var batches []models.Batch
	database.DB.Order("id desc").Limit(10).Find(&batches)
	for _, b := range batches {
		fmt.Printf("ID: %d, ProdID: %d, BranchID: %d, Qty: %d, Num: %s\n", b.ID, b.ProductID, b.BranchID, b.Quantity, b.BatchNumber)
	}

	fmt.Println("\n=== Stock Movements (Latest 10) ===")
	var movements []models.StockMovement
	database.DB.Order("id desc").Limit(10).Find(&movements)
	for _, m := range movements {
		fmt.Printf("ID: %d, ProdID: %d, BranchID: %d, Qty: %d, Type: %s, Ref: %s\n", m.ID, m.ProductID, m.BranchID, m.Quantity, m.Type, m.Reference)
	}
}
