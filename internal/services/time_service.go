package services

import (
	"database/sql"
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
	res, err := s.db.Exec("INSERT INTO employees (first_name, last_name, email) VALUES (?, ?, ?)", emp.FirstName, emp.LastName, emp.Email)
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

func (s *TimeService) CreateProject(proj *models.Project) error {
	res, err := s.db.Exec("INSERT INTO projects (name, company_name, description) VALUES (?, ?, ?)", proj.Name, proj.CompanyName, proj.Description)
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

func (s *TimeService) CreateTimeEntry(te *models.TimeEntry) error {
	res, err := s.db.Exec("INSERT INTO time_entries (assignment_id, billing_code_id, date, hours, task_description) VALUES (?, ?, ?, ?, ?)", te.AssignmentID, te.BillingCodeID, te.Date, te.Hours, te.TaskDescription)
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

func (s *TimeService) GetEmployees() ([]models.Employee, error) {
	rows, err := s.db.Query("SELECT id, first_name, last_name, email FROM employees")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []models.Employee
	for rows.Next() {
		var e models.Employee
		if err := rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Email); err != nil {
			return nil, err
		}
		emps = append(emps, e)
	}
	return emps, nil
}

func (s *TimeService) GetProjects() ([]models.Project, error) {
	rows, err := s.db.Query("SELECT id, name, company_name, description FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projs []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.CompanyName, &p.Description); err != nil {
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
