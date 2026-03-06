package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	LogService *services.LogService
}

func NewCustomerHandler(logService *services.LogService) *CustomerHandler {
	return &CustomerHandler{LogService: logService}
}

type customerInput struct {
	Name    string `json:"name" binding:"required,min=2,max=255"`
	Email   string `json:"email" binding:"omitempty,email"`
	Phone   string `json:"phone" binding:"max=50"`
	Address string `json:"address"`
}

// List returns all customers.
func (h *CustomerHandler) List(c *gin.Context) {
	query := database.DB.Model(&models.Customer{})

	// Search by name or phone
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var customers []models.Customer
	if err := query.Order("name ASC").Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": customers})
}

// GetByID returns a single customer.
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"customer": customer})
}

// Create creates a new customer.
func (h *CustomerHandler) Create(c *gin.Context) {
	var input customerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	customer := models.Customer{
		Name:    input.Name,
		Email:   input.Email,
		Phone:   input.Phone,
		Address: input.Address,
	}

	if err := database.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "CREATE", "Customer", strconv.Itoa(int(customer.ID)), fmt.Sprintf("Created customer: %s", customer.Name), c.ClientIP())
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Customer created", "customer": customer})
}

// Update updates an existing customer.
func (h *CustomerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	var input customerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	customer.Name = input.Name
	customer.Email = input.Email
	customer.Phone = input.Phone
	customer.Address = input.Address

	if err := database.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "UPDATE", "Customer", strconv.Itoa(int(customer.ID)), fmt.Sprintf("Updated customer: %s", customer.Name), c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer updated", "customer": customer})
}

// Delete deletes a customer.
func (h *CustomerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	result := database.DB.Delete(&models.Customer{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	userIDValue, _ := c.Get("userID")
	if userIDValue != nil {
		h.LogService.Record(userIDValue.(uint), "DELETE", "Customer", strconv.Itoa(int(id)), fmt.Sprintf("Deleted customer #%d", id), c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
}
