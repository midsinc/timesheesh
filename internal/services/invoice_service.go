package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/go-pdf/fpdf"
	"os"
	"time"
	"timesheesh/internal/models"
)

type InvoiceService struct {
	db *sql.DB
}

func NewInvoiceService(db *sql.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

type InvoiceData struct {
	ProjectName        string
	CompanyName        string
	EmployeeName       string
	EmployeeAddress    string
	PaymentTargetDate  time.Time
	Entries            []DailyEntry
	TotalHours         float64
	TotalBillable      float64
}

type DailyEntry struct {
	Date        time.Time
	Hours       float64
	UnitPrice   float64
	LineTotal   float64
	Description string
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderInvoiceEntryRow(pdf *fpdf.Fpdf, entry DailyEntry) {
	const (
		dateWidth        = 28.0
		descriptionWidth = 68.0
		hoursWidth       = 24.0
		unitPriceWidth   = 35.0
		lineTotalWidth   = 35.0
		lineHeight       = 5.0
		maxLines         = 2
	)

	x, y := pdf.GetXY()

	pdf.SetFont("Arial", "", 9)
	lines := pdf.SplitLines([]byte(entry.Description), descriptionWidth)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := string(lines[maxLines-1])
		if len(last) > 3 {
			last = last[:len(last)-3] + "..."
		} else {
			last += "..."
		}
		lines[maxLines-1] = []byte(last)
	}
	rowHeight := lineHeight * float64(maxInt(len(lines), 1))

	pdf.Rect(x, y, dateWidth, rowHeight, "")
	pdf.Rect(x+dateWidth, y, descriptionWidth, rowHeight, "")
	pdf.Rect(x+dateWidth+descriptionWidth, y, hoursWidth, rowHeight, "")
	pdf.Rect(x+dateWidth+descriptionWidth+hoursWidth, y, unitPriceWidth, rowHeight, "")
	pdf.Rect(x+dateWidth+descriptionWidth+hoursWidth+unitPriceWidth, y, lineTotalWidth, rowHeight, "")

	pdf.SetXY(x, y)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(dateWidth, rowHeight, entry.Date.Format("2006-01-02"), "", 0, "C", false, 0, "")

	pdf.SetXY(x+dateWidth, y)
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(descriptionWidth, lineHeight, string(bytes.Join(lines, []byte("\n"))), "", "L", false)

	pdf.SetXY(x+dateWidth+descriptionWidth, y)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(hoursWidth, rowHeight, fmt.Sprintf("%.2f", entry.Hours), "", 0, "C", false, 0, "")
	pdf.CellFormat(unitPriceWidth, rowHeight, fmt.Sprintf("$%.2f", entry.UnitPrice), "", 0, "C", false, 0, "")
	pdf.CellFormat(lineTotalWidth, rowHeight, fmt.Sprintf("$%.2f", entry.LineTotal), "", 0, "C", false, 0, "")
	pdf.Ln(rowHeight)
}

func parseStoredDate(value string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time entry date %q", value)
}

func (s *InvoiceService) GenerateMonthlyInvoice(projectID int, employeeID int, year int, month time.Month, descriptionMode string, outputPath string) error {
	pdfData, err := s.GenerateMonthlyInvoicePDF(projectID, employeeID, year, month, descriptionMode)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdfData, 0o644)
}

func (s *InvoiceService) GenerateMonthlyInvoicePDF(projectID int, employeeID int, year int, month time.Month, descriptionMode string) ([]byte, error) {
	data, err := s.BuildInvoiceData(projectID, employeeID, year, month, descriptionMode)
	if err != nil {
		return nil, err
	}

	// 5. Create PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice: %s", data.ProjectName))
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Billing Company: %s", data.CompanyName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Employee: %s", data.EmployeeName))
	pdf.Ln(8)
	pdf.MultiCell(0, 7, fmt.Sprintf("Employee Address: %s", data.EmployeeAddress), "", "L", false)
	pdf.Cell(40, 10, fmt.Sprintf("Period: %s %d", month.String(), year))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Payment Target Date: %s", data.PaymentTargetDate.Format("2006-01-02")))
	pdf.Ln(15)

	// Table Header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(28, 10, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(68, 10, "Task Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(24, 10, "Hours", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 10, "Unit Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 10, "Line Total", "1", 0, "C", true, 0, "")
	pdf.Ln(10)

	// Table Body
	for _, entry := range data.Entries {
		renderInvoiceEntryRow(pdf, entry)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(160, 10, "Total Hours: ")
	pdf.Cell(30, 10, fmt.Sprintf("%.2f", data.TotalHours))
	pdf.Ln(8)
	pdf.Cell(160, 10, "Total Billable: ")
	pdf.Cell(30, 10, fmt.Sprintf("$%.2f", data.TotalBillable))
	pdf.Ln(20)
	pdf.Cell(80, 10, "Employee Signature: ______________________________")

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (s *InvoiceService) BuildInvoiceData(projectID int, employeeID int, year int, month time.Month, descriptionMode string) (InvoiceData, error) {
	if descriptionMode == "" {
		descriptionMode = "task"
	}
	var proj models.Project
	err := s.db.QueryRow("SELECT id, name, company_name, description, default_payment_terms FROM projects WHERE id = ?", projectID).Scan(&proj.ID, &proj.Name, &proj.CompanyName, &proj.Description, &proj.DefaultPaymentTerms)
	if err != nil {
		return InvoiceData{}, fmt.Errorf("project not found: %w", err)
	}

	// 2. Fetch Employee
	var emp models.Employee
	err = s.db.QueryRow("SELECT id, first_name, last_name, address FROM employees WHERE id = ?", employeeID).Scan(&emp.ID, &emp.FirstName, &emp.LastName, &emp.Address)
	if err != nil {
		return InvoiceData{}, fmt.Errorf("employee not found: %w", err)
	}

	// 3. Fetch Assignment for rates
	var asn models.Assignment
	err = s.db.QueryRow("SELECT id, billable_rate FROM assignments WHERE employee_id = ? AND project_id = ?", employeeID, projectID).Scan(&asn.ID, &asn.BillableRate)
	if err != nil {
		return InvoiceData{}, fmt.Errorf("no assignment found for employee on this project: %w", err)
	}

	// 4. Fetch Time Entries for the month
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	rows, err := s.db.Query(`
		SELECT date, hours, task_description
		FROM time_entries
		WHERE assignment_id = ? AND date >= ? AND date < ?
		ORDER BY date ASC`, asn.ID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return InvoiceData{}, err
	}
	defer rows.Close()

	var entries []DailyEntry
	var totalHours float64
	for rows.Next() {
		var d DailyEntry
		var dateStr string
		if err := rows.Scan(&dateStr, &d.Hours, &d.Description); err != nil {
			return InvoiceData{}, err
		}
		d.Date, err = parseStoredDate(dateStr)
		if err != nil {
			return InvoiceData{}, err
		}
		if descriptionMode == "project" {
			d.Description = fmt.Sprintf("%s / %s", proj.Name, proj.Description)
		}
		d.UnitPrice = asn.BillableRate
		d.LineTotal = d.Hours * d.UnitPrice
		entries = append(entries, d)
		totalHours += d.Hours
	}

	if len(entries) == 0 {
		return InvoiceData{}, fmt.Errorf("no time entries found for %s %d", month.String(), year)
	}

	totalBillable := totalHours * asn.BillableRate
	paymentTargetDate := endDate.AddDate(0, 0, proj.DefaultPaymentTerms)

	return InvoiceData{
		ProjectName:        proj.Name,
		CompanyName:        proj.CompanyName,
		EmployeeName:       fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
		EmployeeAddress:    emp.Address,
		PaymentTargetDate:  paymentTargetDate,
		Entries:            entries,
		TotalHours:         totalHours,
		TotalBillable:      totalBillable,
	}, nil
}
