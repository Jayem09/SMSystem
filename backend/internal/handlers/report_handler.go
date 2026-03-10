package handlers

import (
	"net/http"
	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

type AdvisorPerformance struct {
	AdvisorName string `json:"advisor_name"`
	TiresSold   int    `json:"tires_sold"`
}

type CategorySale struct {
	Category   string  `json:"category"`
	TotalSales float64 `json:"total_sales"`
}

type PaymentSummary struct {
	Method string  `json:"method"`
	Total  float64 `json:"total"`
}

type DailySummaryResponse struct {
	Date               string               `json:"date"`
	AdvisorPerformance []AdvisorPerformance `json:"advisor_performance"`
	CategorySales      []CategorySale       `json:"category_sales"`
	PaymentSummary     []PaymentSummary     `json:"payment_summary"`
	AccountReceivables float64              `json:"account_receivables"`
	TotalSales         float64              `json:"total_sales"`
}

func (h *ReportHandler) GetDailySummary(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Parse the date to get the start and end of the day in Local time
	startOfDay, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		startOfDay, _ = time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	}
	endOfDay := startOfDay.Add(24 * time.Hour)

	branchID, _ := c.Get("branchID") // Assuming branchID is set by a middleware

	// 1. Advisor Performance (Tires Sold)
	var advisors []AdvisorPerformance
	database.DB.Table("orders").
		Select("orders.service_advisor_name as advisor_name, SUM(order_items.quantity) as tires_sold").
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Joins("JOIN products ON order_items.product_id = products.id").
		Joins("JOIN categories ON products.category_id = categories.id").
		Where("orders.status = 'completed' AND (categories.name LIKE '%TIRE%' OR categories.name LIKE '%MAGS%')").
		Where("orders.created_at >= ? AND orders.created_at < ?", startOfDay, endOfDay).
		Where("orders.branch_id = ?", branchID).
		Group("orders.service_advisor_name").
		Order("tires_sold DESC").
		Scan(&advisors)

	// 2. Category Sales
	var categories []CategorySale
	database.DB.Table("order_items").
		Select("COALESCE(categories.name, 'Uncategorized') as category, SUM(order_items.subtotal) as total_sales").
		Joins("JOIN products ON order_items.product_id = products.id").
		Joins("LEFT JOIN categories ON products.category_id = categories.id").
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("orders.status = 'completed'").
		Where("orders.created_at >= ? AND orders.created_at < ?", startOfDay, endOfDay).
		Where("orders.branch_id = ?", branchID).
		Group("categories.name").
		Order("total_sales DESC").
		Scan(&categories)

	// 3. Payment Summary
	var payments []PaymentSummary
	database.DB.Model(&models.Order{}).
		Select("payment_method as method, SUM(total_amount) as total").
		Where("status = 'completed'").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Where("branch_id = ?", branchID).
		Group("payment_method").
		Scan(&payments)

	// 4. Account Receivables (Pending Orders)
	var ar float64
	database.DB.Model(&models.Order{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status = 'pending'").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Where("branch_id = ?", branchID).
		Scan(&ar)

	// 5. Total Sales (Completed)
	var totalSales float64
	database.DB.Model(&models.Order{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status = 'completed'").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Where("branch_id = ?", branchID).
		Scan(&totalSales)

	c.JSON(http.StatusOK, DailySummaryResponse{
		Date:               dateStr,
		AdvisorPerformance: advisors,
		CategorySales:      categories,
		PaymentSummary:     payments,
		AccountReceivables: ar,
		TotalSales:         totalSales,
	})
}
