package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	LogService *services.LogService
}

func NewSettingsHandler(logService *services.LogService) *SettingsHandler {
	return &SettingsHandler{LogService: logService}
}


func (h *SettingsHandler) GetAll(c *gin.Context) {
	var settings []models.Setting
	if err := database.DB.Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	result := make(map[string]interface{})
	for _, s := range settings {
		var parsed interface{}
		
		if err := json.Unmarshal([]byte(s.Value), &parsed); err == nil {
			result[s.Key] = parsed
		} else {
			result[s.Key] = s.Value
		}
	}

	c.JSON(http.StatusOK, result)
}


func (h *SettingsHandler) UpdateBulk(c *gin.Context) {
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Force stdout flush
	fmt.Println("[Settings] UpdateBulk called with keys:", getMapKeys(input))
	log.Println("[Settings] UpdateBulk called with keys:", getMapKeys(input))

	tx := database.DB.Begin()
	for key, value := range input {
		strValue := ""
		switch v := value.(type) {
		case string:
			strValue = v
		default:
			bytes, _ := json.Marshal(v)
			strValue = string(bytes)
		}

		fmt.Printf("[Settings] Processing key=%s, value_len=%d\n", key, len(strValue))

		setting := models.Setting{Key: key, Value: strValue}
		var existing models.Setting
		err := tx.First(&existing, "key = ?", key).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("[Settings] Creating NEW key=%s\n", key)
				if err := tx.Create(&setting).Error; err != nil {
					fmt.Printf("[Settings] Create FAILED: %v\n", err)
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Create failed: " + err.Error()})
					return
				}
			} else {
				fmt.Printf("[Settings] Query error: %v\n", err)
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed: " + err.Error()})
				return
			}
		} else {
			fmt.Printf("[Settings] Updating existing key=%s, old_value_len=%d\n", key, len(existing.Value))
			if err := tx.Model(&existing).Update("value", strValue).Error; err != nil {
				fmt.Printf("[Settings] Update FAILED: %v\n", err)
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed: " + err.Error()})
				return
			}
		}
		fmt.Printf("[Settings] Key=%s saved successfully\n", key)
	}
	tx.Commit()
	fmt.Println("[Settings] All settings saved successfully!")

	currentUserID, _ := c.Get("userID")
	if currentUserID != nil {
		h.LogService.Record(currentUserID.(uint), "UPDATE", "Settings", "0", "Updated system settings", c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
