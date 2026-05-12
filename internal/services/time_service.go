package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"timesheesh/internal/models"
)

type TimeService struct {
	db *sql.DB
}

func (s *TimeService) GetDB() *sql.DB {
	return s.db
}

func NewTimeService(db *sql.DB) *TimeService {
	return &TimeService{db: db}
}

func (s *TimeService) CreateEmployee(emp *models.Employee) error {
	res, err := s.db.Exec("INSERT INTO employees (first_name, last_name, email, address) VALUES (?, ?, ?, ?)", emp.FirstName, emp.LastName, emp.Email, emp.Address)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	emp.ID = int(id)
	return nil
}

func (s *TimeService) UpdateEmployee(emp *models.Employee) error {
	_, err := s.db.Exec(
		"UPDATE employees SET first_name = ?, last_name = ?, email = ?, address = ? WHERE id = ?",
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Address,
		emp.ID,
	)
	return err
}

func (s *TimeService) CreateProject(proj *models.Project) error {
	if proj.DefaultPaymentTerms <= 0 {
		proj.DefaultPaymentTerms = 30
	}
	res, err := s.db.Exec(
		"INSERT INTO projects (name, company_name, description, default_payment_terms) VALUES (?, ?, ?, ?)",
		proj.Name,
		proj.CompanyName,
		proj.Description,
		proj.DefaultPaymentTerms,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	proj.ID = int(id)
	return nil
}

func (s *TimeService) UpdateProject(proj *models.Project) error {
	if proj.DefaultPaymentTerms <= 0 {
		proj.DefaultPaymentTerms = 30
	}
	_, err := s.db.Exec(
		"UPDATE projects SET name = ?, company_name = ?, description = ?, default_payment_terms = ? WHERE id = ?",
		proj.Name,
		proj.CompanyName,
		proj.Description,
		proj.DefaultPaymentTerms,
		proj.ID,
	)
	return err
}

func (s *TimeService) CreateAssignment(asn *models.Assignment) error {
	res, err := s.db.Exec("INSERT INTO assignments (employee_id, project_id, billable_rate, pay_rate) VALUES (?, ?, ?, ?)", asn.EmployeeID, asn.ProjectID, asn.BillableRate, asn.PayRate)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	asn.ID = int(id)
	return nil
}

func (s *TimeService) UpdateAssignment(asn *models.Assignment) error {
	_, err := s.db.Exec(
		"UPDATE assignments SET employee_id = ?, project_id = ?, billable_rate = ?, pay_rate = ? WHERE id = ?",
		asn.EmployeeID,
		asn.ProjectID,
		asn.BillableRate,
		asn.PayRate,
		asn.ID,
	)
	return err
}

func (s *TimeService) CreateTimeEntry(te *models.TimeEntry) error {
	var billingCodeID interface{}
	if te.BillingCodeID > 0 {
		billingCodeID = te.BillingCodeID
	}

	res, err := s.db.Exec(
		"INSERT INTO time_entries (assignment_id, billing_code_id, date, hours, task_description) VALUES (?, ?, ?, ?, ?)",
		te.AssignmentID,
		billingCodeID,
		te.Date.Format("2006-01-02"),
		te.Hours,
		te.TaskDescription,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	te.ID = int(id)
	return nil
}

func (s *TimeService) UpdateTimeEntry(te *models.TimeEntry) error {
	var billingCodeID interface{}
	if te.BillingCodeID > 0 {
		billingCodeID = te.BillingCodeID
	}

	_, err := s.db.Exec(
		"UPDATE time_entries SET assignment_id = ?, billing_code_id = ?, date = ?, hours = ?, task_description = ? WHERE id = ?",
		te.AssignmentID,
		billingCodeID,
		te.Date.Format("2006-01-02"),
		te.Hours,
		te.TaskDescription,
		te.ID,
	)
	return err
}

func (s *TimeService) CreateBillingCode(bc *models.BillingCode) error {
	res, err := s.db.Exec("INSERT INTO billing_codes (project_id, code, description) VALUES (?, ?, ?)", bc.ProjectID, bc.Code, bc.Description)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	bc.ID = int(id)
	return nil
}

func (s *TimeService) GetBillingCodesByProject(projectID int) ([]models.BillingCode, error) {
	rows, err := s.db.Query("SELECT id, project_id, code, description FROM billing_codes WHERE project_id = ?", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []models.BillingCode
	for rows.Next() {
		var bc models.BillingCode
		if err := rows.Scan(&bc.ID, &bc.ProjectID, &bc.Code, &bc.Description); err != nil {
			return nil, err
		}
		codes = append(codes, bc)
	}
	return codes, nil
}

func (s *TimeService) GetEmployeeAssignments(employeeID int) ([]models.Assignment, error) {
	rows, err := s.db.Query("SELECT id, employee_id, project_id, billable_rate, pay_rate FROM assignments WHERE employee_id = ?", employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var asns []models.Assignment
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.ProjectID, &a.BillableRate, &a.PayRate); err != nil {
			return nil, err
		}
		asns = append(asns, a)
	}
	return asns, nil
}

func (s *TimeService) GetEmployeeByID(employeeID int) (models.Employee, error) {
	var employee models.Employee
	err := s.db.QueryRow(
		"SELECT id, first_name, last_name, email, address FROM employees WHERE id = ?",
		employeeID,
	).Scan(&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email, &employee.Address)
	if err != nil {
		return models.Employee{}, err
	}
	return employee, nil
}

func (s *TimeService) GetEmployees() ([]models.Employee, error) {
	rows, err := s.db.Query("SELECT id, first_name, last_name, email, address FROM employees")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []models.Employee
	for rows.Next() {
		var e models.Employee
		if err := rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Email, &e.Address); err != nil {
			return nil, err
		}
		emps = append(emps, e)
	}
	return emps, nil
}

func (s *TimeService) GetProjects() ([]models.Project, error) {
	rows, err := s.db.Query("SELECT id, name, company_name, description, default_payment_terms FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projs []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.CompanyName, &p.Description, &p.DefaultPaymentTerms); err != nil {
			return nil, err
		}
		projs = append(projs, p)
	}
	return projs, nil
}

func (s *TimeService) GetAssignments() ([]models.Assignment, error) {
	rows, err := s.db.Query("SELECT id, employee_id, project_id, billable_rate, pay_rate FROM assignments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var asns []models.Assignment
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.ProjectID, &a.BillableRate, &a.PayRate); err != nil {
			return nil, err
		}
		asns = append(asns, a)
	}
	return asns, nil
}

func (s *TimeService) GetProjectsByEmployee(employeeID int) ([]models.AgentProjectView, error) {
	assignments, err := s.GetEmployeeAssignments(employeeID)
	if err != nil {
		return nil, err
	}

	projects, err := s.GetProjects()
	if err != nil {
		return nil, err
	}
	projectByID := make(map[int]models.Project, len(projects))
	for _, project := range projects {
		projectByID[project.ID] = project
	}

	billingCodesByProject, err := s.billingCodeSummariesByProject(projects)
	if err != nil {
		return nil, err
	}

	views := make([]models.AgentProjectView, 0, len(assignments))
	for _, assignment := range assignments {
		project, ok := projectByID[assignment.ProjectID]
		if !ok {
			return nil, fmt.Errorf("project %d not found for assignment %d", assignment.ProjectID, assignment.ID)
		}
		views = append(views, models.AgentProjectView{
			ID:           project.ID,
			Name:         project.Name,
			CompanyName:  project.CompanyName,
			Description:  project.Description,
			AssignmentID: assignment.ID,
			BillingCodes: billingCodesByProject[project.ID],
		})
	}

	return views, nil
}

func (s *TimeService) GetAllProjectsDetailed() ([]models.AgentProjectView, error) {
	projects, err := s.GetProjects()
	if err != nil {
		return nil, err
	}

	billingCodesByProject, err := s.billingCodeSummariesByProject(projects)
	if err != nil {
		return nil, err
	}

	views := make([]models.AgentProjectView, 0, len(projects))
	for _, project := range projects {
		views = append(views, models.AgentProjectView{
			ID:           project.ID,
			Name:         project.Name,
			CompanyName:  project.CompanyName,
			Description:  project.Description,
			BillingCodes: billingCodesByProject[project.ID],
		})
	}

	return views, nil
}

func (s *TimeService) GetAssignmentsDetailedByEmployee(employeeID int) ([]models.AgentAssignmentView, error) {
	assignments, err := s.GetEmployeeAssignments(employeeID)
	if err != nil {
		return nil, err
	}
	return s.assignmentViews(assignments)
}

func (s *TimeService) GetAssignmentsDetailed() ([]models.AgentAssignmentView, error) {
	assignments, err := s.GetAssignments()
	if err != nil {
		return nil, err
	}
	return s.assignmentViews(assignments)
}

func (s *TimeService) assignmentViews(assignments []models.Assignment) ([]models.AgentAssignmentView, error) {
	employees, err := s.GetEmployees()
	if err != nil {
		return nil, err
	}
	projects, err := s.GetProjects()
	if err != nil {
		return nil, err
	}

	employeeByID := make(map[int]models.Employee, len(employees))
	for _, employee := range employees {
		employeeByID[employee.ID] = employee
	}
	projectByID := make(map[int]models.Project, len(projects))
	for _, project := range projects {
		projectByID[project.ID] = project
	}

	billingCodesByProject, err := s.billingCodeSummariesByProject(projects)
	if err != nil {
		return nil, err
	}

	views := make([]models.AgentAssignmentView, 0, len(assignments))
	for _, assignment := range assignments {
		employee, ok := employeeByID[assignment.EmployeeID]
		if !ok {
			return nil, fmt.Errorf("employee %d not found for assignment %d", assignment.EmployeeID, assignment.ID)
		}
		project, ok := projectByID[assignment.ProjectID]
		if !ok {
			return nil, fmt.Errorf("project %d not found for assignment %d", assignment.ProjectID, assignment.ID)
		}

		views = append(views, models.AgentAssignmentView{
			ID:            assignment.ID,
			EmployeeID:    employee.ID,
			EmployeeName:  strings.TrimSpace(employee.FirstName + " " + employee.LastName),
			EmployeeEmail: employee.Email,
			ProjectID:     project.ID,
			ProjectName:   project.Name,
			BillableRate:  assignment.BillableRate,
			PayRate:       assignment.PayRate,
			BillingCodes:  billingCodesByProject[project.ID],
		})
	}

	return views, nil
}

func (s *TimeService) billingCodeSummariesByProject(projects []models.Project) (map[int][]models.AgentBillingCodeSummary, error) {
	summaries := make(map[int][]models.AgentBillingCodeSummary, len(projects))
	for _, project := range projects {
		codes, err := s.GetBillingCodesByProject(project.ID)
		if err != nil {
			return nil, err
		}

		projectCodes := make([]models.AgentBillingCodeSummary, 0, len(codes))
		for _, code := range codes {
			projectCodes = append(projectCodes, models.AgentBillingCodeSummary{
				ID:          code.ID,
				Code:        code.Code,
				Description: code.Description,
			})
		}
		summaries[project.ID] = projectCodes
	}
	return summaries, nil
}

func (s *TimeService) GetTimeEntries(month string) ([]models.TimeEntry, error) {
	query := "SELECT id, assignment_id, billing_code_id, date, hours, task_description FROM time_entries"
	args := []interface{}{}
	if month != "" {
		query += " WHERE strftime('%Y-%m', date) = ?"
		args = append(args, month)
	}
	query += " ORDER BY date DESC, id DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.TimeEntry
	for rows.Next() {
		var te models.TimeEntry
		var rawDate string
		var billingCodeID sql.NullInt64
		if err := rows.Scan(&te.ID, &te.AssignmentID, &billingCodeID, &rawDate, &te.Hours, &te.TaskDescription); err != nil {
			return nil, err
		}
		if billingCodeID.Valid {
			te.BillingCodeID = int(billingCodeID.Int64)
		}
		te.Date, err = time.Parse("2006-01-02", rawDate)
		if err != nil {
			te.Date, err = time.Parse(time.RFC3339, rawDate)
			if err != nil {
				return nil, err
			}
		}
		entries = append(entries, te)
	}
	return entries, nil
}
