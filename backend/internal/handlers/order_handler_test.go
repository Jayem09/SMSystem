package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"smsystem-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func TestCheckoutInputBindsMechanicName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{
		"guest_name": "Walk In",
		"guest_phone": "09171234567",
		"service_advisor_name": "Mike",
		"mechanic_name": "Jun",
		"payment_method": "cash",
		"status": "completed",
		"receipt_type": "SI",
		"business_address": "Lipa City",
		"items": [{"product_id": 1, "quantity": 1}]
	}`)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	context.Request = request

	var input checkoutInput
	if err := context.ShouldBindJSON(&input); err != nil {
		t.Fatalf("expected payload to bind successfully, got %v", err)
	}

	if input.MechanicName != "Jun" {
		t.Fatalf("expected mechanic name Jun, got %q", input.MechanicName)
	}
}

func TestOrderJSONIncludesMechanicName(t *testing.T) {
	payload, err := json.Marshal(models.Order{
		ID:                 42,
		ServiceAdvisorName: "Mike",
		MechanicName:       "Jun",
	})
	if err != nil {
		t.Fatalf("expected order JSON to marshal, got %v", err)
	}

	if !bytes.Contains(payload, []byte(`"mechanic_name":"Jun"`)) {
		t.Fatalf("expected mechanic_name to be present in JSON, got %s", payload)
	}
}

func TestCheckoutInputBindsTintnerAndCarwasherNames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{
		"guest_name": "Walk In",
		"guest_phone": "09171234567",
		"service_advisor_name": "Mike",
		"mechanic_name": "Jun",
		"tintner_name": "Ken",
		"carwasher_name": "Paul",
		"payment_method": "cash",
		"status": "completed",
		"receipt_type": "SI",
		"business_address": "Lipa City",
		"items": [{"product_id": 1, "quantity": 1}]
	}`)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	context.Request = request

	var input checkoutInput
	if err := context.ShouldBindJSON(&input); err != nil {
		t.Fatalf("expected payload to bind successfully, got %v", err)
	}

	if input.TintnerName != "Ken" {
		t.Fatalf("expected tintner name Ken, got %q", input.TintnerName)
	}

	if input.CarwasherName != "Paul" {
		t.Fatalf("expected carwasher name Paul, got %q", input.CarwasherName)
	}
}

func TestOrderJSONIncludesTintnerAndCarwasherNames(t *testing.T) {
	payload, err := json.Marshal(models.Order{
		ID:                 42,
		ServiceAdvisorName: "Mike",
		MechanicName:       "Jun",
		TintnerName:        "Ken",
		CarwasherName:      "Paul",
	})
	if err != nil {
		t.Fatalf("expected order JSON to marshal, got %v", err)
	}

	if !bytes.Contains(payload, []byte(`"tintner_name":"Ken"`)) {
		t.Fatalf("expected tintner_name to be present in JSON, got %s", payload)
	}

	if !bytes.Contains(payload, []byte(`"carwasher_name":"Paul"`)) {
		t.Fatalf("expected carwasher_name to be present in JSON, got %s", payload)
	}
}
