package handlers

import (
	"testing"
	"time"
)

func TestBuildTransactionSearchFilterIncludesMechanic(t *testing.T) {
	clause, args := buildTransactionSearchFilter("  Jomar  ")

	expectedClause := "COALESCE(customers.name, orders.guest_name, '') LIKE ? OR COALESCE(orders.plate_number, '') LIKE ? OR products.name LIKE ? OR orders.service_advisor_name LIKE ? OR orders.mechanic_name LIKE ?"
	if clause != expectedClause {
		t.Fatalf("expected clause %q, got %q", expectedClause, clause)
	}

	if len(args) != 5 {
		t.Fatalf("expected 5 search args, got %d", len(args))
	}

	for _, arg := range args {
		if arg != "%Jomar%" {
			t.Fatalf("expected all args to be %%Jomar%%, got %#v", args)
		}
	}
}

func TestBuildTransactionRowsIncludesMechanicName(t *testing.T) {
	rows := buildTransactionRows([]transactionRawRow{ {
		CreatedAt:          time.Date(2026, 4, 29, 9, 30, 0, 0, time.UTC),
		ID:                 42,
		ReceiptType:        "SI",
		BranchName:         "LIPA A",
		CustomerName:       "John Doe",
		ServiceAdvisorName: "Joel",
		MechanicName:       "Jomar",
		ItemName:           "Accelera Tire",
		UnitOfMeasure:      "pc",
		CategoryName:       "Tires",
		Quantity:           2,
		UnitPrice:          1029,
		Subtotal:           2058,
		PaymentMethod:      "cash",
		OrderStatus:        "completed",
	} })

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	if rows[0].ServiceAdvisorName != "Joel" {
		t.Fatalf("expected advisor Joel, got %q", rows[0].ServiceAdvisorName)
	}

	if rows[0].MechanicName != "Jomar" {
		t.Fatalf("expected mechanic Jomar, got %q", rows[0].MechanicName)
	}
}
