package main

import (
	"fmt"
	"log"
	"os"
	"smsystem-backend/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load .env from backend directory
	_ = godotenv.Load(".env")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("=== Products and Stock Cache ===")
	var products []models.Product
	db.Find(&products)
	for _, p := range products {
		fmt.Printf("ID: %d | Name: %s | Stock Cache: %d\n", p.ID, p.Name, p.Stock)
	}

	fmt.Println("\n=== Batches Table (Raw) ===")
	var batches []models.Batch
	db.Find(&batches)
	for _, b := range batches {
		fmt.Printf("ID: %d | ProductID: %d | BranchID: %d | Qty: %d\n", b.ID, b.ProductID, b.BranchID, b.Quantity)
	}

	fmt.Println("\n=== Stock Sum per Product (Global) ===")
	type StockSum struct {
		ProductID uint
		Total     int
	}
	var sums []StockSum
	db.Table("batches").Select("product_id, SUM(quantity) as total").Group("product_id").Scan(&sums)
	for _, s := range sums {
		fmt.Printf("ProductID: %d | Total Batch Qty: %d\n", s.ProductID, s.Total)
	}

	fmt.Println("\n=== Recent Stock Movements (Last 10) ===")
	var movements []models.StockMovement
	db.Order("created_at DESC").Limit(10).Find(&movements)
	for _, m := range movements {
		fmt.Printf("ID: %d | ProductID: %d | Type: %s | Qty: %d | Ref: %s\n", m.ID, m.ProductID, m.Type, m.Quantity, m.Reference)
	}
}
