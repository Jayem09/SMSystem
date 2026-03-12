package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	LogService *services.LogService
}

func NewOrderHandler(logSvc *services.LogService) *OrderHandler {
	return &OrderHandler{LogService: logSvc}
}

type orderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type orderInput struct {
	CustomerID         *uint            `json:"customer_id"`
	GuestName          string           `json:"guest_name"`
	GuestPhone         string           `json:"guest_phone"`
	ServiceAdvisorName string           `json:"service_advisor_name"`
	PaymentMethod      string           `json:"payment_method" binding:"required"`
	DiscountAmount     float64          `json:"discount_amount"`
	DiscountType       string           `json:"discount_type"` // "fixed" or "percentage"
	TaxAmount          float64          `json:"tax_amount"`
	IsTaxInclusive     bool             `json:"is_tax_inclusive"`
	Items              []orderItemInput `json:"items" binding:"required,min=1,dive"`
}

type statusInput struct {
	Status string `json:"status" binding:"required,oneof=pending confirmed completed cancelled"`
}

type checkoutInput struct {
	orderInput
	Status             string  `json:"status" binding:"omitempty,oneof=pending completed"`
	ReceiptType        string  `json:"receipt_type" binding:"required,oneof=SI DR"`
	TIN                string  `json:"tin"`
	BusinessAddress    string  `json:"business_address"`
	WithholdingTaxRate float64 `json:"withholding_tax_rate"`
}

// List returns all orders with their items, customer, and user.
func (h *OrderHandler) List(c *gin.Context) {
	branchID, _ := c.Get("branchID")
	query := database.DB.Where("branch_id = ?", branchID).Preload("Customer").Preload("User").Preload("Items.Product")

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by customer
	if customerID := c.Query("customer_id"); customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}

	var orders []models.Order
	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetByID returns a single order with full details.
func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.Order
	if err := database.DB.Preload("Customer").Preload("User").Preload("Items.Product").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order})
}

// Create creates a new order with items inside a database transaction.
// It validates stock, calculates totals, and reduces product stock atomically (unless status is pending).
func (h *OrderHandler) Create(c *gin.Context) {
	var input checkoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Default status to completed if not provided
	orderStatus := "completed"
	if input.Status != "" {
		orderStatus = input.Status
	}

	// Get the authenticated user ID and branch ID
	userID, exists := c.Get("userID")
	branchID, _ := c.Get("branchID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Verify customer exists if ID provided
	if input.CustomerID != nil && *input.CustomerID > 0 {
		var customer models.Customer
		if err := database.DB.First(&customer, *input.CustomerID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not found"})
			return
		}
	}

	// Run everything in a transaction
	var order models.Order
	order.BranchID = branchID.(uint)
	uID := userID.(uint)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var totalAmount float64
		var orderItems []models.OrderItem

		// First pass: validate and calculate
		for _, item := range input.Items {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return errors.New("product ID " + strconv.Itoa(int(item.ProductID)) + " not found")
			}

			if orderStatus == "completed" && !product.IsService {
				if product.Stock < item.Quantity {
					return errors.New("insufficient stock for " + product.Name + " (available: " + strconv.Itoa(product.Stock) + ")")
				}
			}

			subtotal := product.Price * float64(item.Quantity)
			totalAmount += subtotal
			orderItems = append(orderItems, models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitPrice: product.Price,
				Subtotal:  subtotal,
			})
		}

		// Calculate final total
		finalTotal := totalAmount
		if input.DiscountType == "percentage" {
			finalTotal -= totalAmount * (input.DiscountAmount / 100)
		} else {
			finalTotal -= input.DiscountAmount
		}
		if !input.IsTaxInclusive {
			finalTotal += input.TaxAmount
		}

		// Create the order first to get an ID
		order = models.Order{
			CustomerID:         input.CustomerID,
			GuestName:          input.GuestName,
			GuestPhone:         input.GuestPhone,
			UserID:             uID,
			ServiceAdvisorName: input.ServiceAdvisorName,
			TotalAmount:        finalTotal,
			DiscountAmount:     input.DiscountAmount,
			DiscountType:       input.DiscountType,
			TaxAmount:          input.TaxAmount,
			IsTaxInclusive:     input.IsTaxInclusive,
			Status:             orderStatus,
			PaymentMethod:      input.PaymentMethod,
			Items:              orderItems,
			ReceiptType:        input.ReceiptType,
			TIN:                input.TIN,
			BusinessAddress:    input.BusinessAddress,
			WithholdingTaxRate: input.WithholdingTaxRate,
			BranchID:           order.BranchID,
		}

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %v", err)
		}

		// Second pass: Deduct stock if completed
		if orderStatus == "completed" {
			for _, item := range orderItems {
				var product models.Product
				tx.First(&product, item.ProductID)
				if product.IsService {
					continue
				}

				remainingToDeduct := item.Quantity
				var batches []models.Batch
				batchQuery := tx.Where("product_id = ? AND quantity > 0", product.ID)
				if order.BranchID != 0 {
					batchQuery = batchQuery.Where("branch_id = ?", order.BranchID)
				}
				
				if err := batchQuery.Order("expiry_date ASC, created_at ASC").Find(&batches).Error; err != nil {
					return fmt.Errorf("failed to find available batches for %s", product.Name)
				}

				for i := range batches {
					if remainingToDeduct <= 0 {
						break
					}
					deduct := remainingToDeduct
					if batches[i].Quantity < deduct {
						deduct = batches[i].Quantity
					}

					tx.Model(&batches[i]).Update("quantity", batches[i].Quantity - deduct)

					movement := models.StockMovement{
						ProductID:   product.ID,
						BatchID:     &batches[i].ID,
						WarehouseID: batches[i].WarehouseID,
						BranchID:    batches[i].BranchID,
						UserID:      &uID,
						Type:        models.MovementTypeOut,
						Quantity:    -deduct,
						Reference:   fmt.Sprintf("POS Order #%d", order.ID),
					}
					tx.Create(&movement)
					remainingToDeduct -= deduct
				}

				// SELF-HEALING: If we still need to deduct but have no more batches,
				// check if the product has legacy cached stock.
				if remainingToDeduct > 0 && product.Stock >= remainingToDeduct {
					fmt.Printf("🔧 Self-healing: Creating legacy batch for product %d to satisfy order #%d\n", product.ID, order.ID)
					var warehouse models.Warehouse
					whQuery := tx.Model(&models.Warehouse{})
					if order.BranchID != 0 {
						whQuery = whQuery.Where("branch_id = ?", order.BranchID)
					}
					if err := whQuery.First(&warehouse).Error; err == nil {
						legacyBatch := models.Batch{
							ProductID:   product.ID,
							WarehouseID: warehouse.ID,
							BranchID:    warehouse.BranchID,
							BatchNumber: "LEGACY-SYNC",
							Quantity:    product.Stock - remainingToDeduct, // Initialize with remaining legacy stock
						}
						tx.Create(&legacyBatch)
						
						deduct := remainingToDeduct
						movement := models.StockMovement{
							ProductID:   product.ID,
							BatchID:     &legacyBatch.ID,
							WarehouseID: warehouse.ID,
							BranchID:    warehouse.BranchID,
							UserID:      &uID,
							Type:        models.MovementTypeOut,
							Quantity:    -deduct,
							Reference:   fmt.Sprintf("POS Order #%d (Legacy Auto-sync)", order.ID),
						}
						tx.Create(&movement)
						remainingToDeduct -= deduct
					}
				}

				if remainingToDeduct > 0 {
					return fmt.Errorf("insufficient batch stock for %s during final deduction", product.Name)
				}

				// Update cache
				tx.Model(&product).Update("stock", product.Stock-item.Quantity)
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reload with all relationships
	database.DB.Preload("Customer").Preload("User").Preload("Items.Product").First(&order, order.ID)

	h.LogService.Record(userID.(uint), "CREATE", "Order", strconv.Itoa(int(order.ID)), fmt.Sprintf("Checked out POS Order #%d", order.ID), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"message": "Order created", "order": order})
}

// UpdateStatus updates an order's status.
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var input statusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	oldStatus := order.Status

	// If transitioning from pending to completed, deduct stock
	if oldStatus == "pending" && input.Status == "completed" {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			// Fetch items
			var items []models.OrderItem
			if err := tx.Where("order_id = ?", id).Find(&items).Error; err != nil {
				return err
			}

			userIDValue, _ := c.Get("userID")
			uID := userIDValue.(uint)

			for _, item := range items {
				var product models.Product
				if err := tx.First(&product, item.ProductID).Error; err != nil {
					return errors.New("product ID " + strconv.Itoa(int(item.ProductID)) + " not found")
				}

				if !product.IsService {
					// FIFO Deduction from Batches
					remainingToDeduct := item.Quantity
					var batches []models.Batch
					
					// Deduct from batches belonging to this branch
					batchQuery := tx.Where("product_id = ? AND quantity > 0", product.ID)
					if order.BranchID != 0 {
						batchQuery = batchQuery.Where("branch_id = ?", order.BranchID)
					}
					
					if err := batchQuery.Order("expiry_date ASC, created_at ASC").Find(&batches).Error; err != nil {
						return fmt.Errorf("failed to find available batches for %s", product.Name)
					}

					for i := range batches {
						if remainingToDeduct <= 0 {
							break
						}

						deductFromBatch := remainingToDeduct
						if batches[i].Quantity < deductFromBatch {
							deductFromBatch = batches[i].Quantity
						}

						// Update batch
						if err := tx.Model(&batches[i]).Update("quantity", batches[i].Quantity - deductFromBatch).Error; err != nil {
							return fmt.Errorf("failed to update batch for %s", product.Name)
						}

						// Record Stock Movement
						movement := models.StockMovement{
							ProductID:   product.ID,
							BatchID:     &batches[i].ID,
							WarehouseID: batches[i].WarehouseID,
							BranchID:    batches[i].BranchID,
							UserID:      &uID,
							Type:        models.MovementTypeOut,
							Quantity:    -deductFromBatch,
							Reference:   fmt.Sprintf("POS Order #%d (Completed Pending)", order.ID),
						}

						if err := tx.Create(&movement).Error; err != nil {
							return fmt.Errorf("failed to record movement for %s", product.Name)
						}

						remainingToDeduct -= deductFromBatch
					}

					// SELF-HEALING: Legacy sync
					if remainingToDeduct > 0 && product.Stock >= remainingToDeduct {
						var warehouse models.Warehouse
						whQuery := tx.Model(&models.Warehouse{})
						if order.BranchID != 0 {
							whQuery = whQuery.Where("branch_id = ?", order.BranchID)
						}
						if err := whQuery.First(&warehouse).Error; err == nil {
							legacyBatch := models.Batch{
								ProductID:   product.ID,
								WarehouseID: warehouse.ID,
								BranchID:    warehouse.BranchID,
								BatchNumber: "LEGACY-SYNC",
								Quantity:    product.Stock - remainingToDeduct,
							}
							tx.Create(&legacyBatch)
							
							deduct := remainingToDeduct
							movement := models.StockMovement{
								ProductID:   product.ID,
								BatchID:     &legacyBatch.ID,
								WarehouseID: warehouse.ID,
								BranchID:    warehouse.BranchID,
								UserID:      &uID,
								Type:        models.MovementTypeOut,
								Quantity:    -deduct,
								Reference:   fmt.Sprintf("POS Order #%d (Pending Legacy Sync)", order.ID),
							}
							tx.Create(&movement)
							remainingToDeduct -= deduct
						}
					}

					if remainingToDeduct > 0 {
						return fmt.Errorf("insufficient batch stock for %s (need %d more)", product.Name, remainingToDeduct)
					}

					// Update product cache
					if err := tx.Model(&product).Update("stock", product.Stock-item.Quantity).Error; err != nil {
						return errors.New("failed to update product stock cache for " + product.Name)
					}
				}
			}

			order.Status = input.Status
			return tx.Save(&order).Error
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to complete order", "details": err.Error()})
			return
		}
	} else {
		// Just update the status normally (e.g. to cancelled)
		order.Status = input.Status
		if err := database.DB.Save(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}
	}

	database.DB.Preload("Customer").Preload("User").Preload("Items.Product").First(&order, order.ID)
	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "UPDATE_STATUS", "Order", strconv.Itoa(int(order.ID)), fmt.Sprintf("Status changed from %s to %s", oldStatus, input.Status), c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated", "order": order})
}

// Delete deletes an order and its items.
func (h *OrderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Delete items first, then order in a transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Order{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "DELETE", "Order", strconv.Itoa(int(id)), fmt.Sprintf("Deleted order #%d", id), c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
