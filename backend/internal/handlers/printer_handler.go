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

type PrinterHandler struct {
	PrinterService *services.PrinterService
	LogService     *services.LogService
}

func NewPrinterHandler(printerSvc *services.PrinterService, logSvc *services.LogService) *PrinterHandler {
	return &PrinterHandler{
		PrinterService: printerSvc,
		LogService:     logSvc,
	}
}

type printInput struct {
	PrinterName string `json:"printer_name" binding:"required"`
}

func (h *PrinterHandler) PrintSI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var input printInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Printer name is required"})
		return
	}

	var order models.Order
	if err := database.DB.Preload("Customer").Preload("Branch").Preload("Items.Product").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	data, err := h.PrinterService.GenerateSI(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate SI: %v", err)})
		return
	}

	if err := h.PrinterService.PrintRaw(input.PrinterName, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Printing failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sales Invoice sent to printer"})
}

func (h *PrinterHandler) PrintDR(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var input printInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Printer name is required"})
		return
	}

	var order models.Order
	if err := database.DB.Preload("Customer").Preload("Branch").Preload("Items.Product").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	data, err := h.PrinterService.GenerateDR(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate DR: %v", err)})
		return
	}

	if err := h.PrinterService.PrintRaw(input.PrinterName, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Printing failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery Receipt sent to printer"})
}

func (h *PrinterHandler) ListPrinters(c *gin.Context) {
	printers, err := h.PrinterService.ListPrinters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list printers: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"printers": printers})
}
