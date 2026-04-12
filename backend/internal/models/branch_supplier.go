package models

import (
	"time"
)

type BranchSupplier struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	BranchID   uint      `gorm:"uniqueIndex:idx_branch_supplier;not null" json:"branch_id"`
	SupplierID uint      `gorm:"uniqueIndex:idx_branch_supplier;not null" json:"supplier_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BranchSupplier) TableName() string {
	return "branch_suppliers"
}
