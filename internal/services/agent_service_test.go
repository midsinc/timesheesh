package services

import (
	"path/filepath"
	"testing"

	"timesheesh/internal/db"
	"timesheesh/internal/models"
)

func TestGetProjectsByEmployeeIncludesBillingCodes(t *testing.T) {
	timeSvc := newAgentTestService(t)
	employee, project, assignment := seedAgentTestRecords(t, timeSvc)

	billingCode := &models.BillingCode{
		ProjectID:   project.ID,
		Code:        "DEV-01",
		Description: "Development",
	}
	if err := timeSvc.CreateBillingCode(billingCode); err != nil {
		t.Fatalf("create billing code: %v", err)
	}

	projects, err := timeSvc.GetProjectsByEmployee(employee.ID)
	if err != nil {
		t.Fatalf("get projects by employee: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	view := projects[0]
	if view.AssignmentID != assignment.ID {
		t.Fatalf("expected assignment id %d, got %d", assignment.ID, view.AssignmentID)
	}
	if len(view.BillingCodes) != 1 || view.BillingCodes[0].Code != "DEV-01" {
		t.Fatalf("expected billing code DEV-01, got %+v", view.BillingCodes)
	}
}

func TestGetAssignmentsDetailedByEmployeeIncludesProjectAndEmployeeFields(t *testing.T) {
	timeSvc := newAgentTestService(t)
	employee, project, assignment := seedAgentTestRecords(t, timeSvc)

	assignments, err := timeSvc.GetAssignmentsDetailedByEmployee(employee.ID)
	if err != nil {
		t.Fatalf("get assignments detailed by employee: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}

	view := assignments[0]
	if view.ID != assignment.ID {
		t.Fatalf("expected assignment id %d, got %d", assignment.ID, view.ID)
	}
	if view.EmployeeEmail != employee.Email {
		t.Fatalf("expected employee email %q, got %q", employee.Email, view.EmployeeEmail)
	}
	if view.ProjectName != project.Name {
		t.Fatalf("expected project name %q, got %q", project.Name, view.ProjectName)
	}
}

func TestGetAssignmentsDetailedReturnsAllAssignments(t *testing.T) {
	timeSvc := newAgentTestService(t)
	seedAgentTestRecords(t, timeSvc)

	secondEmployee := &models.Employee{
		FirstName: "Grace",
		LastName:  "Hopper",
		Email:     "grace@example.com",
		Address:   "456 Example St",
	}
	if err := timeSvc.CreateEmployee(secondEmployee); err != nil {
		t.Fatalf("create second employee: %v", err)
	}

	secondProject := &models.Project{
		Name:                "Zeus",
		CompanyName:         "Globex",
		Description:         "Second project",
		DefaultPaymentTerms: 45,
	}
	if err := timeSvc.CreateProject(secondProject); err != nil {
		t.Fatalf("create second project: %v", err)
	}

	secondAssignment := &models.Assignment{
		EmployeeID:   secondEmployee.ID,
		ProjectID:    secondProject.ID,
		BillableRate: 200,
		PayRate:      120,
	}
	if err := timeSvc.CreateAssignment(secondAssignment); err != nil {
		t.Fatalf("create second assignment: %v", err)
	}

	assignments, err := timeSvc.GetAssignmentsDetailed()
	if err != nil {
		t.Fatalf("get assignments detailed: %v", err)
	}
	if len(assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(assignments))
	}
}

func newAgentTestService(t *testing.T) *TimeService {
	t.Helper()

	database, err := db.InitDB(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	return NewTimeService(database)
}

func seedAgentTestRecords(t *testing.T, timeSvc *TimeService) (*models.Employee, *models.Project, *models.Assignment) {
	t.Helper()

	employee := &models.Employee{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Address:   "123 Example St",
	}
	if err := timeSvc.CreateEmployee(employee); err != nil {
		t.Fatalf("create employee: %v", err)
	}

	project := &models.Project{
		Name:                "Apollo",
		CompanyName:         "Acme Corp",
		Description:         "CLI test",
		DefaultPaymentTerms: 30,
	}
	if err := timeSvc.CreateProject(project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	assignment := &models.Assignment{
		EmployeeID:   employee.ID,
		ProjectID:    project.ID,
		BillableRate: 150,
		PayRate:      90,
	}
	if err := timeSvc.CreateAssignment(assignment); err != nil {
		t.Fatalf("create assignment: %v", err)
	}

	return employee, project, assignment
}
