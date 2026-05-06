package models

import "time"

type Employee struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}

type Project struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	CompanyName         string `json:"company_name"`
	Description         string `json:"description"`
	DefaultPaymentTerms int    `json:"default_payment_terms"`
}

type BillingCode struct {
	ID          int    `json:"id"`
	ProjectID   int    `json:"project_id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type Assignment struct {
	ID           int     `json:"id"`
	EmployeeID   int     `json:"employee_id"`
	ProjectID    int     `json:"project_id"`
	BillableRate float64 `json:"billable_rate"`
	PayRate      float64 `json:"pay_rate"`
}

type TimeEntry struct {
	ID              int       `json:"id"`
	AssignmentID    int       `json:"assignment_id"`
	BillingCodeID   int       `json:"billing_code_id"`
	Date            time.Time `json:"date"`
	Hours           float64   `json:"hours"`
	TaskDescription string    `json:"task_description"`
}
