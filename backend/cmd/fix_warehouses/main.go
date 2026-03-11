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

	for _, b := range branches {
		var count int64
		database.DB.Model(&models.Warehouse{}).Where("branch_id = ?", b.ID).Count(&count)
		if count == 0 {
			fmt.Printf("Creating default warehouse for branch: %s (ID: %d)\n", b.Name, b.ID)
			w := models.Warehouse{
				Name:     "Main Warehouse",
				BranchID: b.ID,
				Address:  "Default Location",
			}
			database.DB.Create(&w)
		} else {
			fmt.Printf("Branch %s (ID: %d) already has %d warehouses\n", b.Name, b.ID, count)
		}
	}
	fmt.Println("Done!")
}
