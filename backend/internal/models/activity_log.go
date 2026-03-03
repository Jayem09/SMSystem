package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Action    string         `gorm:"size:100;not null" json:"action"` // e.g., "DELETE", "UPDATE_PRICE", "LOGIN"
	Entity    string         `gorm:"size:100;not null" json:"entity"` // e.g., "Product", "Order"
	EntityID  string         `gorm:"size:50" json:"entity_id"`        // ID of the affected resource
	Details   string         `gorm:"type:text" json:"details"`        // JSON or stringified changes
	IPAddress string         `gorm:"size:45" json:"ip_address"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
