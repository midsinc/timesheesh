package services

import (
	"path/filepath"
	"testing"
	"time"

	"timesheesh/internal/db"
	"timesheesh/internal/models"
)

func TestBuildInvoiceDataIncludesPaymentTermsAddressAndLineTotals(t *testing.T) {
	database, err := db.InitDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	timeSvc := NewTimeService(database)
	invoiceSvc := NewInvoiceService(database)

	employee := &models.Employee{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Address:   "123 Example St, Indianapolis, IN",
	}
	if err := timeSvc.CreateEmployee(employee); err != nil {
		t.Fatalf("create employee: %v", err)
	}

	project := &models.Project{
		Name:                "Apollo",
		CompanyName:         "Acme Corp",
		Description:         "Invoice test",
		DefaultPaymentTerms: 15,
	}
	if err := timeSvc.CreateProject(project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	assignment := &models.Assignment{
		EmployeeID:   employee.ID,
		ProjectID:    project.ID,
		BillableRate: 125,
		PayRate:      80,
	}
	if err := timeSvc.CreateAssignment(assignment); err != nil {
		t.Fatalf("create assignment: %v", err)
	}

	entry := &models.TimeEntry{
		AssignmentID:    assignment.ID,
		Date:            time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
		Hours:           8,
		TaskDescription: "Implementation",
	}
	if err := timeSvc.CreateTimeEntry(entry); err != nil {
		t.Fatalf("create time entry: %v", err)
	}

	data, err := invoiceSvc.BuildInvoiceData(project.ID, employee.ID, 2026, time.April, "task")
	if err != nil {
		t.Fatalf("build invoice data: %v", err)
	}

	if data.EmployeeAddress != employee.Address {
		t.Fatalf("expected employee address %q, got %q", employee.Address, data.EmployeeAddress)
	}
	expectedTargetDate := time.Date(2026, time.May, 16, 0, 0, 0, 0, time.UTC)
	if !data.PaymentTargetDate.Equal(expectedTargetDate) {
		t.Fatalf("expected payment target %s, got %s", expectedTargetDate, data.PaymentTargetDate)
	}
	if len(data.Entries) != 1 {
		t.Fatalf("expected 1 invoice entry, got %d", len(data.Entries))
	}
	if data.Entries[0].UnitPrice != 125 || data.Entries[0].LineTotal != 1000 {
		t.Fatalf("unexpected line item pricing: %+v", data.Entries[0])
	}

	projectModeData, err := invoiceSvc.BuildInvoiceData(project.ID, employee.ID, 2026, time.April, "project")
	if err != nil {
		t.Fatalf("build project mode invoice data: %v", err)
	}
	if projectModeData.Entries[0].Description != "Apollo / Invoice test" {
		t.Fatalf("expected project description mode, got %q", projectModeData.Entries[0].Description)
	}
}
