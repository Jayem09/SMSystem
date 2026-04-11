package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"smsystem-backend/internal/database"
)

type OllamaClient struct {
	BaseURL string
	Model   string
}

func NewOllamaClient() *OllamaClient {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   "llama3.2:1b",
	}
}

type OllamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func (o *OllamaClient) GetBusinessContext(branchID uint) string {
	db := database.DB

	type salesSum struct{ Total float64 }
	type orderCount struct{ Count int }
	type productCount struct{ Count int }
	type customerCount struct{ Count int }
	type lowStockCount struct{ Count int }

	var todaysSales, monthSales, totalSales salesSum
	var todaysOrders, monthOrders, totalOrders orderCount
	var totalProducts productCount
	var totalCustomers customerCount
	var lowStock lowStockCount

	db.Raw("SELECT COALESCE(SUM(total_amount - discount_amount), 0) as total FROM orders WHERE DATE(created_at) = CURDATE() AND status != 'cancelled'").Scan(&todaysSales)
	db.Raw("SELECT COALESCE(SUM(total_amount - discount_amount), 0) as total FROM orders WHERE YEAR(created_at) = YEAR(NOW()) AND MONTH(created_at) = MONTH(NOW()) AND status != 'cancelled'").Scan(&monthSales)
	db.Raw("SELECT COALESCE(SUM(total_amount - discount_amount), 0) as total FROM orders WHERE status != 'cancelled'").Scan(&totalSales)

	db.Raw("SELECT COUNT(*) as count FROM orders WHERE DATE(created_at) = CURDATE() AND status != 'cancelled'").Scan(&todaysOrders)
	db.Raw("SELECT COUNT(*) as count FROM orders WHERE YEAR(created_at) = YEAR(NOW()) AND MONTH(created_at) = MONTH(NOW()) AND status != 'cancelled'").Scan(&monthOrders)
	db.Raw("SELECT COUNT(*) as count FROM orders WHERE status != 'cancelled'").Scan(&totalOrders)

	db.Raw("SELECT COUNT(*) as count FROM products WHERE deleted_at IS NULL").Scan(&totalProducts)
	db.Raw("SELECT COUNT(*) as count FROM customers").Scan(&totalCustomers)
	db.Raw("SELECT COUNT(*) as count FROM products WHERE stock <= reorder_level AND deleted_at IS NULL").Scan(&lowStock)

	return fmt.Sprintf(`Current Business Metrics:
• Today's Sales: ₱%.2f
• This Month's Sales: ₱%.2f
• Total Sales: ₱%.2f
• Today's Orders: %d
• This Month's Orders: %d
• Total Orders: %d
• Total Products: %d
• Total Customers: %d
• Low Stock Items: %d`, todaysSales.Total, monthSales.Total, totalSales.Total, todaysOrders.Count, monthOrders.Count, totalOrders.Count, totalProducts.Count, totalCustomers.Count, lowStock.Count)
}

func (o *OllamaClient) GenerateWithQuestion(prompt string, businessContext string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are a business analytics assistant for an SMS (Sales Management System) for a tire shop.
You help users understand their sales, inventory, customers, and business performance.
Use the following REAL data to answer questions - do NOT make up numbers:
%s

Keep answers brief and actionable. Use bullet points when listing items. Focus on actionable insights.`, businessContext)

	reqBody := OllamaRequest{
		Model: o.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(o.BaseURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Message.Content, nil
}
