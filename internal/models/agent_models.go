package models

type AgentEmployeeSummary struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email"`
}

type AgentBillingCodeSummary struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type AgentProjectView struct {
	ID           int                       `json:"id"`
	Name         string                    `json:"name"`
	CompanyName  string                    `json:"company_name"`
	Description  string                    `json:"description"`
	AssignmentID int                       `json:"assignment_id,omitempty"`
	BillingCodes []AgentBillingCodeSummary `json:"billing_codes"`
}

type AgentAssignmentView struct {
	ID            int                       `json:"id"`
	EmployeeID    int                       `json:"employee_id"`
	EmployeeName  string                    `json:"employee_name,omitempty"`
	EmployeeEmail string                    `json:"employee_email,omitempty"`
	ProjectID     int                       `json:"project_id"`
	ProjectName   string                    `json:"project_name"`
	BillableRate  float64                   `json:"billable_rate"`
	PayRate       float64                   `json:"pay_rate"`
	BillingCodes  []AgentBillingCodeSummary `json:"billing_codes"`
}

type AgentProjectsResponse struct {
	Employee *AgentEmployeeSummary `json:"employee,omitempty"`
	Projects []AgentProjectView    `json:"projects"`
}

type AgentAssignmentsResponse struct {
	Employee    *AgentEmployeeSummary `json:"employee,omitempty"`
	Assignments []AgentAssignmentView `json:"assignments"`
}
