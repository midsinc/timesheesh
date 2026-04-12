package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"timesheesh/internal/db"
	"timesheesh/internal/models"
	"timesheesh/internal/services"
	"timesheesh/internal/web"

	"github.com/spf13/cobra"
)

var (
	dbPath  string
	service *services.TimeService
	invSvc  *services.InvoiceService
)

var rootCmd = &cobra.Command{
	Use:   "timesheesh",
	Short: "Time Sheesh - employee time tracking and invoicing",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "timesheesh.db", "Path to SQLite database")
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

// --- Employee Commands ---

var addEmployeeCmd = &cobra.Command{
	Use:   "add-employee [first] [last] [email]",
	Short: "Add a new employee",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		emp := &models.Employee{FirstName: args[0], LastName: args[1], Email: args[2]}
		if err := service.CreateEmployee(emp); err != nil {
			log.Fatalf("Error creating employee: %v", err)
		}
		fmt.Printf("Employee %s %s added with ID %d\n", emp.FirstName, emp.LastName, emp.ID)
	},
}

var listEmployeesCmd = &cobra.Command{
	Use:   "list-employees",
	Short: "List all employees",
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		emps, err := service.GetEmployees()
		if err != nil {
			log.Fatalf("Error getting employees: %v", err)
		}
		fmt.Println("Employees:")
		for _, e := range emps {
			fmt.Printf("[%d] %s %s (%s)\n", e.ID, e.FirstName, e.LastName, e.Email)
		}
	},
}

// --- Project Commands ---

var addProjectCmd = &cobra.Command{
	Use:   "add-project [name] [company] [description]",
	Short: "Add a new project",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		proj := &models.Project{Name: args[0], CompanyName: args[1], Description: args[2]}
		if err := service.CreateProject(proj); err != nil {
			log.Fatalf("Error creating project: %v", err)
		}
		fmt.Printf("Project %s added with ID %d\n", proj.Name, proj.ID)
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		projs, err := service.GetProjects()
		if err != nil {
			log.Fatalf("Error getting projects: %v", err)
		}
		fmt.Println("Projects:")
		for _, p := range projs {
			fmt.Printf("[%d] %s (%s)\n", p.ID, p.Name, p.CompanyName)
		}
	},
}

// --- Assignment Commands ---

var addAssignmentCmd = &cobra.Command{
	Use:   "add-assignment [empID] [projID] [billRate] [payRate]",
	Short: "Assign employee to project with rates",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		eID, _ := strconv.Atoi(args[0])
		pID, _ := strconv.Atoi(args[1])
		bRate, _ := strconv.ParseFloat(args[2], 64)
		pRate, _ := strconv.ParseFloat(args[3], 64)

		asn := &models.Assignment{EmployeeID: eID, ProjectID: pID, BillableRate: bRate, PayRate: pRate}
		if err := service.CreateAssignment(asn); err != nil {
			log.Fatalf("Error creating assignment: %v", err)
		}
		fmt.Printf("Employee %d assigned to project %d with billable rate %.2f\n", eID, pID, bRate)
	},
}

// --- Time Entry Commands ---

var addTimeCmd = &cobra.Command{
	Use:   "add-time [asnID] [date:YYYY-MM-DD] [hours] [desc]",
	Short: "Add time entry for an assignment",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		aID, _ := strconv.Atoi(args[0])
		date, err := time.Parse("2006-01-02", args[1])
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}
		hrs, _ := strconv.ParseFloat(args[2], 64)

		te := &models.TimeEntry{AssignmentID: aID, Date: date, Hours: hrs, TaskDescription: args[3]}
		if err := service.CreateTimeEntry(te); err != nil {
			log.Fatalf("Error creating time entry: %v", err)
		}
		fmt.Println("Time entry added successfully.")
	},
}

// --- Invoice Command ---

var genInvoiceCmd = &cobra.Command{
	Use:   "invoice [projID] [empID] [year] [month] [output]",
	Short: "Generate monthly PDF invoice",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
		pID, _ := strconv.Atoi(args[0])
		eID, _ := strconv.Atoi(args[1])
		yr, _ := strconv.Atoi(args[2])
		mVal, _ := strconv.Atoi(args[3])
		out := args[4]

		if err := invSvc.GenerateMonthlyInvoice(pID, eID, yr, time.Month(mVal), out); err != nil {
			log.Fatalf("Error generating invoice: %v", err)
		}
		fmt.Printf("Invoice successfully generated at %s\n", out)
	},
}

// --- Web Server Command ---

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
	// Emp
	empCmd := &cobra.Command{Use: "employee", Short: "Employee management"}
	empCmd.AddCommand(addEmployeeCmd, listEmployeesCmd)
	rootCmd.AddCommand(empCmd)

	// Proj
	projCmd := &cobra.Command{Use: "project", Short: "Project management"}
	projCmd.AddCommand(addProjectCmd, listProjectsCmd)
	rootCmd.AddCommand(projCmd)

	// Asgn
	asgnCmd := &cobra.Command{Use: "assignment", Short: "Assignment management"}
	asgnCmd.AddCommand(addAssignmentCmd)
	rootCmd.AddCommand(asgnCmd)

	// Time
	timeCmd := &cobra.Command{Use: "time", Short: "Time tracking"}
	timeCmd.AddCommand(addTimeCmd)
	rootCmd.AddCommand(timeCmd)

	// Invoice
	rootCmd.AddCommand(genInvoiceCmd)

	// Server
	rootCmd.AddCommand(serverCmd)
}
