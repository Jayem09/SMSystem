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
	fmt.Println("=== LAST ORDER TRACER ===")

	var order models.Order
	if err := database.DB.Preload("Items").Order("id DESC").First(&order).Error; err != nil {
		fmt.Println("No orders found!")
		return
	}

	fmt.Printf("Last Order ID: %d | Status: %s | Total: %.2f\n", order.ID, order.Status, order.TotalAmount)
	for _, item := range order.Items {
		var p models.Product
		var batches []models.Batch
		var totalQty int
		
		database.DB.First(&p, item.ProductID)
		database.DB.Where("product_id = ?", p.ID).Find(&batches)
		for _, b := range batches {
			totalQty += b.Quantity
			fmt.Printf("  -> Batch %d for %s (Branch %d) Qty: %d\n", b.ID, p.Name, b.BranchID, b.Quantity)
		}

		fmt.Printf("  Item purchased: %s x%d | Current calculated stock: %d\n", p.Name, item.Quantity, totalQty)
	}

	var logs []models.ActivityLog
	database.DB.Order("id DESC").Limit(3).Find(&logs)
	fmt.Println("\n=== LAST 3 ACTIVITY LOGS ===")
	for _, l := range logs {
		fmt.Printf("Log ID: %d | Action: %s | Details: %s\n", l.ID, l.Action, l.Details)
	}
}
