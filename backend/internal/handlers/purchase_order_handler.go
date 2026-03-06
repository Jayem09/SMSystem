package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type PurchaseOrderHandler struct {
	LogService *services.LogService
}

func NewPurchaseOrderHandler(logService *services.LogService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{LogService: logService}
}

type purchaseOrderItemInput struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitCost  float64 `json:"unit_cost" binding:"required,min=0"`
}

type purchaseOrderInput struct {
	SupplierID uint                     `json:"supplier_id" binding:"required"`
	OrderDate  string                   `json:"order_date" binding:"required"`
	Notes      string                   `json:"notes"`
	Items      []purchaseOrderItemInput `json:"items" binding:"required,min=1,dive"`
}

// List returns all purchase orders with supplier and items.
func (h *PurchaseOrderHandler) List(c *gin.Context) {
	var orders []models.PurchaseOrder
	if err := database.DB.
		Preload("Supplier").
		Preload("User").
		Preload("Items.Product").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"purchase_orders": orders})
}

// GetByID returns a single purchase order with all details.
func (h *PurchaseOrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase order ID"})
		return
	}

	var po models.PurchaseOrder
	if err := database.DB.
		Preload("Supplier").
		Preload("User").
		Preload("Items.Product").
		First(&po, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"purchase_order": po})
}

// Create creates a new purchase order with items.
func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var input purchaseOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Parse order date
	orderDate, err := time.Parse("2006-01-02", input.OrderDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order date format. Use YYYY-MM-DD"})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("userID")

	// Build items and calculate total
	var totalCost float64
	var items []models.PurchaseOrderItem
	for _, item := range input.Items {
		subtotal := float64(item.Quantity) * item.UnitCost
		totalCost += subtotal
		items = append(items, models.PurchaseOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
			Subtotal:  subtotal,
		})
	}

	po := models.PurchaseOrder{
		SupplierID: input.SupplierID,
		UserID:     userID.(uint),
		Status:     "pending",
		TotalCost:  totalCost,
		OrderDate:  orderDate,
		Notes:      input.Notes,
		Items:      items,
	}

	if err := database.DB.Create(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
		return
	}

	// Reload with relations
	database.DB.Preload("Supplier").Preload("User").Preload("Items.Product").First(&po, po.ID)

	h.LogService.Record(userID.(uint), "CREATE", "Purchase Order", strconv.Itoa(int(po.ID)), fmt.Sprintf("Created PO to supplier #%d", po.SupplierID), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"message": "Purchase order created", "purchase_order": po})
}

// Receive marks a purchase order as received and updates product stock.
func (h *PurchaseOrderHandler) Receive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase order ID"})
		return
	}

	var po models.PurchaseOrder
	if err := database.DB.Preload("Items").First(&po, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	if po.Status == "received" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase order already received"})
		return
	}

	if po.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot receive a cancelled purchase order"})
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Update stock for each item and update cost price
	for _, item := range po.Items {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Updates(map[string]interface{}{
				"stock":      database.DB.Raw("stock + ?", item.Quantity),
				"cost_price": item.UnitCost,
			}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	// Mark PO as received
	now := time.Now()
	po.Status = "received"
	po.ReceivedDate = &now
	if err := tx.Save(&po).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update purchase order"})
		return
	}

	tx.Commit()

	// Reload with relations
	database.DB.Preload("Supplier").Preload("User").Preload("Items.Product").First(&po, po.ID)

	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "UPDATE", "Purchase Order", strconv.Itoa(int(po.ID)), "Received items and updated stock", c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase order received. Stock updated.", "purchase_order": po})
}

// Delete deletes a purchase order (only if pending).
func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase order ID"})
		return
	}

	var po models.PurchaseOrder
	if err := database.DB.First(&po, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	if po.Status == "received" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a received purchase order"})
		return
	}

	// Delete items first, then the PO
	database.DB.Where("purchase_order_id = ?", id).Delete(&models.PurchaseOrderItem{})
	database.DB.Delete(&po)

	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "DELETE", "Purchase Order", strconv.Itoa(int(id)), fmt.Sprintf("Deleted PO #%d", id), c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase order deleted"})
}
