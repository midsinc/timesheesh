package services

import (
	"fmt"
	"timesheesh/internal/models"
	"github.com/go-pdf/fpdf"
	"database/sql"
	"time"
)

type InvoiceService struct {
	db *sql.DB
}

func NewInvoiceService(db *sql.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

type InvoiceData struct {
	ProjectName   string
	CompanyName   string
	EmployeeName  string
	Entries       []DailyEntry
	TotalHours    float64
	TotalBillable float64
}

type DailyEntry struct {
	Date        time.Time
	Hours       float64
	Description string
}

func (s *InvoiceService) GenerateMonthlyInvoice(projectID int, employeeID int, year int, month time.Month, outputPath string) error {
	// 1. Fetch Project and Company
	var proj models.Project
	err := s.db.QueryRow("SELECT id, name, company_name FROM projects WHERE id = ?", projectID).Scan(&proj.ID, &proj.Name, &proj.CompanyName)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// 2. Fetch Employee
	var emp models.Employee
	err = s.db.QueryRow("SELECT id, first_name, last_name FROM employees WHERE id = ?", employeeID).Scan(&emp.ID, &emp.FirstName, &emp.LastName)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	// 3. Fetch Assignment for rates
	var asn models.Assignment
	err = s.db.QueryRow("SELECT id, billable_rate FROM assignments WHERE employee_id = ? AND project_id = ?", employeeID, projectID).Scan(&asn.ID, &asn.BillableRate)
	if err != nil {
		return fmt.Errorf("no assignment found for employee on this project: %w", err)
	}

	// 4. Fetch Time Entries for the month
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	rows, err := s.db.Query(`
		SELECT date, hours, task_description
		FROM time_entries
		WHERE assignment_id = ? AND date >= ? AND date < ?
		ORDER BY date ASC`, asn.ID, startDate, endDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	var entries []DailyEntry
	var totalHours float64
	for rows.Next() {
		var d DailyEntry
		var dateStr string
		if err := rows.Scan(&dateStr, &d.Hours, &d.Description); err != nil {
			return err
		}
		d.Date, _ = time.Parse("2006-01-02", dateStr)
		entries = append(entries, d)
		totalHours += d.Hours
	}

	totalBillable := totalHours * asn.BillableRate

	// 5. Create PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice: %s", proj.Name))
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Billing Company: %s", proj.CompanyName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Employee: %s %s", emp.FirstName, emp.LastName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Period: %s %d", month.String(), year))
	pdf.Ln(15)

	// Table Header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 10, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(120, 10, "Task Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "Hours", "1", 0, "C", true, 0, "")
	pdf.Ln(10)

	// Table Body
	pdf.SetFont("Arial", "", 12)
	for _, entry := range entries {
		pdf.CellFormat(40, 10, entry.Date.Format("2006-01-02"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(120, 10, entry.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%.2f", entry.Hours), "1", 0, "C", false, 0, "")
		pdf.Ln(10)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(160, 10, "Total Hours: ")
	pdf.Cell(30, 10, fmt.Sprintf("%.2f", totalHours))
	pdf.Ln(8)
	pdf.Cell(160, 10, "Total Billable: ")
	pdf.Cell(30, 10, fmt.Sprintf("$%.2f", totalBillable))

	return pdf.OutputFileAndClose(outputPath)
}
