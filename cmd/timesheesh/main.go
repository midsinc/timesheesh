package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"timesheesh/internal/db"
	"timesheesh/internal/models"
	"timesheesh/internal/services"
	"timesheesh/internal/web"

	"github.com/spf13/cobra"
)

var (
	dbPath     string
	jsonOutput bool
	service    *services.TimeService
	invSvc     *services.InvoiceService
)

var rootCmd = &cobra.Command{
	Use:   "timesheesh",
	Short: "Time Sheesh - employee time tracking and invoicing",
	Long: "Time Sheesh provides both a web app and a CLI over the same data model.\n" +
		"The CLI supports create, list, and update flows for the same core records exposed in the UI.\n" +
		"Use --json on any command when a machine-readable response is needed.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "timesheesh.db", "Path to SQLite database")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON instead of human-formatted text")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupDB() {
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	service = services.NewTimeService(database)
	invSvc = services.NewInvoiceService(database)
}

func printOutput(v interface{}, human string) {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			log.Fatalf("Error encoding JSON output: %v", err)
		}
		return
	}
	fmt.Println(human)
}

func mustAtoi(value string, field string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Invalid %s %q: %v", field, value, err)
	}
	return n
}

func mustParseFloat(value string, field string) float64 {
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("Invalid %s %q: %v", field, value, err)
	}
	return n
}

func mustParseDate(value string) time.Time {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		log.Fatalf("Invalid date %q: %v", value, err)
	}
	return date
}

func normalizeRef(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func isNumericRef(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

func resolveEmployeeRef(ref string) int {
	if isNumericRef(ref) {
		return mustAtoi(ref, "employee id")
	}

	employees, err := service.GetEmployees()
	if err != nil {
		log.Fatalf("Error getting employees: %v", err)
	}

	needle := normalizeRef(ref)
	matches := make([]models.Employee, 0)
	for _, emp := range employees {
		fullName := normalizeRef(strings.TrimSpace(emp.FirstName + " " + emp.LastName))
		email := normalizeRef(emp.Email)
		if fullName == needle || email == needle {
			matches = append(matches, emp)
		}
	}

	if len(matches) == 0 {
		log.Fatalf("Employee reference %q not found. Run './timesheesh employee list'.", ref)
	}
	if len(matches) > 1 {
		log.Fatalf("Employee reference %q matched multiple employees. Use email or ID.", ref)
	}
	return matches[0].ID
}

func resolveProjectRef(ref string) int {
	if isNumericRef(ref) {
		return mustAtoi(ref, "project id")
	}

	projects, err := service.GetProjects()
	if err != nil {
		log.Fatalf("Error getting projects: %v", err)
	}

	needle := normalizeRef(ref)
	matches := make([]models.Project, 0)
	for _, proj := range projects {
		if normalizeRef(proj.Name) == needle {
			matches = append(matches, proj)
		}
	}

	if len(matches) == 0 {
		log.Fatalf("Project reference %q not found. Run './timesheesh project list'.", ref)
	}
	if len(matches) > 1 {
		log.Fatalf("Project reference %q matched multiple projects. Use ID.", ref)
	}
	return matches[0].ID
}

func resolveAssignmentRef(ref string) int {
	if isNumericRef(ref) {
		return mustAtoi(ref, "assignment id")
	}

	parts := strings.SplitN(ref, "::", 2)
	if len(parts) != 2 {
		log.Fatalf("Invalid assignment reference %q. Use assignment ID or employee::project, for example 'Ada Lovelace::Apollo'.", ref)
	}

	employeeID := resolveEmployeeRef(parts[0])
	projectID := resolveProjectRef(parts[1])

	assignments, err := service.GetAssignments()
	if err != nil {
		log.Fatalf("Error getting assignments: %v", err)
	}

	for _, assignment := range assignments {
		if assignment.EmployeeID == employeeID && assignment.ProjectID == projectID {
			return assignment.ID
		}
	}

	log.Fatalf("Assignment reference %q not found. Run './timesheesh assignment list'.", ref)
	return 0
}

var addEmployeeCmd = &cobra.Command{
	Use:   "add [first] [last] [email] [address]",
	Short: "Add a new employee",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		emp := &models.Employee{FirstName: args[0], LastName: args[1], Email: args[2], Address: args[3]}
		if err := service.CreateEmployee(emp); err != nil {
			log.Fatalf("Error creating employee: %v", err)
		}
		printOutput(emp, fmt.Sprintf("Employee %s %s added with ID %d", emp.FirstName, emp.LastName, emp.ID))
	},
}

var updateEmployeeCmd = &cobra.Command{
	Use:   "update [employee_ref] [first] [last] [email] [address]",
	Short: "Update an existing employee",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		emp := &models.Employee{
			ID:        resolveEmployeeRef(args[0]),
			FirstName: args[1],
			LastName:  args[2],
			Email:     args[3],
			Address:   args[4],
		}
		if err := service.UpdateEmployee(emp); err != nil {
			log.Fatalf("Error updating employee: %v", err)
		}
		printOutput(emp, fmt.Sprintf("Employee %d updated", emp.ID))
	},
}

var listEmployeesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all employees",
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		emps, err := service.GetEmployees()
		if err != nil {
			log.Fatalf("Error getting employees: %v", err)
		}
		if jsonOutput {
			printOutput(emps, "")
			return
		}
		fmt.Println("Employees:")
		for _, e := range emps {
			fmt.Printf("[%d] %s %s | %s | %s\n", e.ID, e.FirstName, e.LastName, e.Email, e.Address)
		}
	},
}

var addProjectCmd = &cobra.Command{
	Use:   "add [name] [company] [description] [invoice_due_days]",
	Short: "Add a new project",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		proj := &models.Project{
			Name:                args[0],
			CompanyName:         args[1],
			Description:         args[2],
			DefaultPaymentTerms: mustAtoi(args[3], "invoice due days"),
		}
		if err := service.CreateProject(proj); err != nil {
			log.Fatalf("Error creating project: %v", err)
		}
		printOutput(proj, fmt.Sprintf("Project %s added with ID %d", proj.Name, proj.ID))
	},
}

var updateProjectCmd = &cobra.Command{
	Use:   "update [project_ref] [name] [company] [description] [invoice_due_days]",
	Short: "Update an existing project",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		proj := &models.Project{
			ID:                  resolveProjectRef(args[0]),
			Name:                args[1],
			CompanyName:         args[2],
			Description:         args[3],
			DefaultPaymentTerms: mustAtoi(args[4], "invoice due days"),
		}
		if err := service.UpdateProject(proj); err != nil {
			log.Fatalf("Error updating project: %v", err)
		}
		printOutput(proj, fmt.Sprintf("Project %d updated", proj.ID))
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		projs, err := service.GetProjects()
		if err != nil {
			log.Fatalf("Error getting projects: %v", err)
		}
		if jsonOutput {
			printOutput(projs, "")
			return
		}
		fmt.Println("Projects:")
		for _, p := range projs {
			fmt.Printf("[%d] %s | %s | due in %d days\n", p.ID, p.Name, p.CompanyName, p.DefaultPaymentTerms)
		}
	},
}

var addAssignmentCmd = &cobra.Command{
	Use:   "add [employee_ref] [project_ref] [bill_rate] [pay_rate]",
	Short: "Assign an employee to a project",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		asn := &models.Assignment{
			EmployeeID:   resolveEmployeeRef(args[0]),
			ProjectID:    resolveProjectRef(args[1]),
			BillableRate: mustParseFloat(args[2], "bill rate"),
			PayRate:      mustParseFloat(args[3], "pay rate"),
		}
		if err := service.CreateAssignment(asn); err != nil {
			log.Fatalf("Error creating assignment: %v", err)
		}
		printOutput(asn, fmt.Sprintf("Assignment %d created", asn.ID))
	},
}

var updateAssignmentCmd = &cobra.Command{
	Use:   "update [assignment_ref] [employee_ref] [project_ref] [bill_rate] [pay_rate]",
	Short: "Update an assignment",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		asn := &models.Assignment{
			ID:           resolveAssignmentRef(args[0]),
			EmployeeID:   resolveEmployeeRef(args[1]),
			ProjectID:    resolveProjectRef(args[2]),
			BillableRate: mustParseFloat(args[3], "bill rate"),
			PayRate:      mustParseFloat(args[4], "pay rate"),
		}
		if err := service.UpdateAssignment(asn); err != nil {
			log.Fatalf("Error updating assignment: %v", err)
		}
		printOutput(asn, fmt.Sprintf("Assignment %d updated", asn.ID))
	},
}

var listAssignmentsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all assignments",
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		asns, err := service.GetAssignments()
		if err != nil {
			log.Fatalf("Error getting assignments: %v", err)
		}
		if jsonOutput {
			printOutput(asns, "")
			return
		}

		employees, err := service.GetEmployees()
		if err != nil {
			log.Fatalf("Error getting employees: %v", err)
		}
		projects, err := service.GetProjects()
		if err != nil {
			log.Fatalf("Error getting projects: %v", err)
		}

		employeeNames := make(map[int]string, len(employees))
		for _, e := range employees {
			employeeNames[e.ID] = strings.TrimSpace(e.FirstName + " " + e.LastName)
		}
		projectNames := make(map[int]string, len(projects))
		for _, p := range projects {
			projectNames[p.ID] = p.Name
		}

		fmt.Println("Assignments:")
		for _, a := range asns {
			fmt.Printf("[%d] %s :: %s | bill=%.2f pay=%.2f\n",
				a.ID,
				employeeNames[a.EmployeeID],
				projectNames[a.ProjectID],
				a.BillableRate,
				a.PayRate,
			)
		}
	},
}

var addBillingCodeCmd = &cobra.Command{
	Use:   "add [project_ref] [code] [description]",
	Short: "Add a billing code to a project",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		bc := &models.BillingCode{
			ProjectID:   resolveProjectRef(args[0]),
			Code:        args[1],
			Description: args[2],
		}
		if err := service.CreateBillingCode(bc); err != nil {
			log.Fatalf("Error creating billing code: %v", err)
		}
		printOutput(bc, fmt.Sprintf("Billing code %s created with ID %d", bc.Code, bc.ID))
	},
}

var listBillingCodesCmd = &cobra.Command{
	Use:   "list [project_ref]",
	Short: "List billing codes for a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		projectID := resolveProjectRef(args[0])
		codes, err := service.GetBillingCodesByProject(projectID)
		if err != nil {
			log.Fatalf("Error getting billing codes: %v", err)
		}
		if jsonOutput {
			printOutput(codes, "")
			return
		}
		fmt.Printf("Billing Codes for Project %d:\n", projectID)
		for _, bc := range codes {
			fmt.Printf("[%d] %s | %s\n", bc.ID, bc.Code, bc.Description)
		}
	},
}

var addTimeCmd = &cobra.Command{
	Use:   "add [assignment_ref] [date:YYYY-MM-DD] [hours] [task_description] [billing_code_id_optional]",
	Short: "Add a time entry",
	Args:  cobra.RangeArgs(4, 5),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		te := &models.TimeEntry{
			AssignmentID:    resolveAssignmentRef(args[0]),
			Date:            mustParseDate(args[1]),
			Hours:           mustParseFloat(args[2], "hours"),
			TaskDescription: args[3],
		}
		if len(args) == 5 {
			te.BillingCodeID = mustAtoi(args[4], "billing code id")
		}
		if err := service.CreateTimeEntry(te); err != nil {
			log.Fatalf("Error creating time entry: %v", err)
		}
		printOutput(te, fmt.Sprintf("Time entry %d created", te.ID))
	},
}

var updateTimeCmd = &cobra.Command{
	Use:   "update [id] [assignment_ref] [date:YYYY-MM-DD] [hours] [task_description] [billing_code_id_optional]",
	Short: "Update a time entry",
	Args:  cobra.RangeArgs(5, 6),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		te := &models.TimeEntry{
			ID:              mustAtoi(args[0], "time entry id"),
			AssignmentID:    resolveAssignmentRef(args[1]),
			Date:            mustParseDate(args[2]),
			Hours:           mustParseFloat(args[3], "hours"),
			TaskDescription: args[4],
		}
		if len(args) == 6 {
			te.BillingCodeID = mustAtoi(args[5], "billing code id")
		}
		if err := service.UpdateTimeEntry(te); err != nil {
			log.Fatalf("Error updating time entry: %v", err)
		}
		printOutput(te, fmt.Sprintf("Time entry %d updated", te.ID))
	},
}

var listTimeCmd = &cobra.Command{
	Use:   "list [month_optional_YYYY-MM]",
	Short: "List time entries, optionally filtered by month",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		month := ""
		if len(args) == 1 {
			month = args[0]
		}
		entries, err := service.GetTimeEntries(month)
		if err != nil {
			log.Fatalf("Error getting time entries: %v", err)
		}
		if jsonOutput {
			printOutput(entries, "")
			return
		}

		assignments, err := service.GetAssignments()
		if err != nil {
			log.Fatalf("Error getting assignments: %v", err)
		}
		employees, err := service.GetEmployees()
		if err != nil {
			log.Fatalf("Error getting employees: %v", err)
		}
		projects, err := service.GetProjects()
		if err != nil {
			log.Fatalf("Error getting projects: %v", err)
		}

		assignmentLabels := make(map[int]string, len(assignments))
		employeeNames := make(map[int]string, len(employees))
		for _, e := range employees {
			employeeNames[e.ID] = strings.TrimSpace(e.FirstName + " " + e.LastName)
		}
		projectNames := make(map[int]string, len(projects))
		for _, p := range projects {
			projectNames[p.ID] = p.Name
		}
		for _, a := range assignments {
			assignmentLabels[a.ID] = fmt.Sprintf("%s :: %s", employeeNames[a.EmployeeID], projectNames[a.ProjectID])
		}

		fmt.Println("Time Entries:")
		for _, te := range entries {
			fmt.Printf("[%d] assignment=%s billing_code=%d date=%s hours=%.2f desc=%s\n",
				te.ID, assignmentLabels[te.AssignmentID], te.BillingCodeID, te.Date.Format("2006-01-02"), te.Hours, te.TaskDescription)
		}
	},
}

var invoiceDescMode string

var genInvoiceCmd = &cobra.Command{
	Use:   "invoice [project_ref] [employee_ref] [year] [month] [output]",
	Short: "Generate a monthly PDF invoice",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		pID := resolveProjectRef(args[0])
		eID := resolveEmployeeRef(args[1])
		yr := mustAtoi(args[2], "year")
		mVal := mustAtoi(args[3], "month")
		out := args[4]

		if err := invSvc.GenerateMonthlyInvoice(pID, eID, yr, time.Month(mVal), invoiceDescMode, out); err != nil {
			log.Fatalf("Error generating invoice: %v", err)
		}
		printOutput(map[string]string{"output": out, "description_mode": invoiceDescMode}, fmt.Sprintf("Invoice successfully generated at %s", out))
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web application server",
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		ws := web.NewWebServer(service, invSvc)
		fmt.Println("Starting server on :8888...")
		if err := ws.Start("8888"); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

func init() {
	genInvoiceCmd.Flags().StringVar(&invoiceDescMode, "description-mode", "task", "Invoice description mode: task or project")

	empCmd := &cobra.Command{
		Use:   "employee",
		Short: "Create, update, and list employees",
		Long:  "Employee records mirror the web UI and include first name, last name, email, and address.",
	}
	empCmd.AddCommand(addEmployeeCmd, updateEmployeeCmd, listEmployeesCmd)
	rootCmd.AddCommand(empCmd)

	projCmd := &cobra.Command{
		Use:   "project",
		Short: "Create, update, and list projects",
		Long:  "Project records include company, description, and invoice due days used to calculate payment target dates.",
	}
	projCmd.AddCommand(addProjectCmd, updateProjectCmd, listProjectsCmd)
	rootCmd.AddCommand(projCmd)

	asgnCmd := &cobra.Command{
		Use:   "assignment",
		Short: "Create, update, and list assignments",
		Long:  "Assignments connect an employee to a project and define bill and pay rates. Use employee full name/email and project name instead of raw IDs when convenient.",
	}
	asgnCmd.AddCommand(addAssignmentCmd, updateAssignmentCmd, listAssignmentsCmd)
	rootCmd.AddCommand(asgnCmd)

	bcCmd := &cobra.Command{
		Use:   "billing-code",
		Short: "Create and list project billing codes",
		Long:  "Billing codes attach codes and descriptions to projects for invoice and time-entry workflows. Project references can be project names or IDs.",
	}
	bcCmd.AddCommand(addBillingCodeCmd, listBillingCodesCmd)
	rootCmd.AddCommand(bcCmd)

	timeCmd := &cobra.Command{
		Use:   "time",
		Short: "Create, update, and list time entries",
		Long:  "Time entries support optional billing code linkage, matching the web app behavior. Assignment references can be IDs or employee::project pairs.",
	}
	timeCmd.AddCommand(addTimeCmd, updateTimeCmd, listTimeCmd)
	rootCmd.AddCommand(timeCmd)

	rootCmd.AddCommand(genInvoiceCmd)
	rootCmd.AddCommand(serverCmd)
}
