package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"timesheesh/internal/models"
	"timesheesh/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type WebServer struct {
	timeSvc *services.TimeService
	invSvc  *services.InvoiceService
	router  *gin.Engine
}

func normalizeStoredDate(value string) string {
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return value
}

func NewWebServer(timeSvc *services.TimeService, invSvc *services.InvoiceService) *WebServer {
	r := gin.Default()
	s := &WebServer{
		timeSvc: timeSvc,
		invSvc:  invSvc,
		router:  r,
	}

	s.setupRoutes()
	return s
}

func (s *WebServer) setupRoutes() {
	api := s.router.Group("/api")
	{
		api.GET("/employees", s.handleGetEmployees)
		api.POST("/employees", s.handleCreateEmployee)
		api.PUT("/employees/:id", s.handleUpdateEmployee)
		api.GET("/projects", s.handleGetProjects)
		api.POST("/projects", s.handleCreateProject)
		api.PUT("/projects/:id", s.handleUpdateProject)
		api.GET("/assignments", s.handleGetAssignments)
		api.GET("/assignments/employee/:empId", s.handleGetEmployeeAssignments)
		api.POST("/assignments", s.handleCreateAssignment)
		api.PUT("/assignments/:id", s.handleUpdateAssignment)
		api.GET("/billing-codes/:projectId", s.handleGetBillingCodes)
		api.POST("/billing-codes", s.handleCreateBillingCode)
		api.GET("/time", s.handleGetTimeEntries)
		api.POST("/time", s.handleCreateTimeEntry)
		api.PUT("/time/:id", s.handleUpdateTimeEntry)
		api.GET("/invoice", s.handleGenerateInvoice)
	}

	s.router.StaticFile("/", "./static/index.html")
	s.router.Static("/static", "./static")
}

func (s *WebServer) handleGetEmployees(c *gin.Context) {
	emps, err := s.timeSvc.GetEmployees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emps)
}

func (s *WebServer) handleCreateEmployee(c *gin.Context) {
	var emp models.Employee
	if err := c.ShouldBindJSON(&emp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.timeSvc.CreateEmployee(&emp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emp)
}

func (s *WebServer) handleUpdateEmployee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee id"})
		return
	}
	var emp models.Employee
	if err := c.ShouldBindJSON(&emp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	emp.ID = id
	if err := s.timeSvc.UpdateEmployee(&emp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emp)
}

func (s *WebServer) handleGetProjects(c *gin.Context) {
	projs, err := s.timeSvc.GetProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projs)
}

func (s *WebServer) handleCreateProject(c *gin.Context) {
	var proj models.Project
	if err := c.ShouldBindJSON(&proj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.timeSvc.CreateProject(&proj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proj)
}

func (s *WebServer) handleUpdateProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	var proj models.Project
	if err := c.ShouldBindJSON(&proj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	proj.ID = id
	if err := s.timeSvc.UpdateProject(&proj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proj)
}

func (s *WebServer) handleGetAssignments(c *gin.Context) {
	asns, err := s.timeSvc.GetAssignments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asns)
}

func (s *WebServer) handleGetEmployeeAssignments(c *gin.Context) {
	empID, _ := strconv.Atoi(c.Param("empId"))
	asns, err := s.timeSvc.GetEmployeeAssignments(empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asns)
}

func (s *WebServer) handleCreateAssignment(c *gin.Context) {
	var asn models.Assignment
	if err := c.ShouldBindJSON(&asn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.timeSvc.CreateAssignment(&asn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asn)
}

func (s *WebServer) handleUpdateAssignment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}
	var asn models.Assignment
	if err := c.ShouldBindJSON(&asn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asn.ID = id
	if err := s.timeSvc.UpdateAssignment(&asn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asn)
}

func (s *WebServer) handleGetBillingCodes(c *gin.Context) {
	projID, _ := strconv.Atoi(c.Param("projectId"))
	codes, err := s.timeSvc.GetBillingCodesByProject(projID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, codes)
}

func (s *WebServer) handleCreateBillingCode(c *gin.Context) {
	var bc models.BillingCode
	if err := c.ShouldBindJSON(&bc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.timeSvc.CreateBillingCode(&bc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bc)
}

func (s *WebServer) handleGetTimeEntries(c *gin.Context) {
	month := c.Query("month")
	query := "SELECT id, assignment_id, billing_code_id, date, hours, task_description FROM time_entries"
	args := []interface{}{}
	if month != "" {
		query += " WHERE strftime('%Y-%m', date) = ?"
		args = append(args, month)
	}
	query += " ORDER BY date DESC, id DESC"

	rows, err := s.timeSvc.GetDB().Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []map[string]interface{}
	for rows.Next() {
		var id, asnID int
		var billingCodeID sql.NullInt64
		var rawDate string
		var hrs float64
		var desc string
		if err := rows.Scan(&id, &asnID, &billingCodeID, &rawDate, &hrs, &desc); err != nil {
			continue
		}
		date := normalizeStoredDate(rawDate)
		var billingCodeValue interface{}
		if billingCodeID.Valid {
			billingCodeValue = billingCodeID.Int64
		}
		entries = append(entries, map[string]interface{}{
			"id": id, "assignment_id": asnID, "billing_code_id": billingCodeValue, "date": date, "hours": hrs, "task_description": desc,
		})
	}
	c.JSON(http.StatusOK, entries)
}

func (s *WebServer) handleCreateTimeEntry(c *gin.Context) {
	te, err := parseTimeEntryRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.timeSvc.CreateTimeEntry(&te); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, te)
}

func (s *WebServer) handleUpdateTimeEntry(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time entry id"})
		return
	}
	te, err := parseTimeEntryRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	te.ID = id
	if err := s.timeSvc.UpdateTimeEntry(&te); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, te)
}

func parseTimeEntryRequest(c *gin.Context) (models.TimeEntry, error) {
	var req struct {
		AssignmentID   int      `json:"assignment_id"`
		BillingCodeID  *int     `json:"billing_code_id"`
		Date           string   `json:"date"`
		Hours          float64  `json:"hours"`
		TaskDescription string  `json:"task_description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return models.TimeEntry{}, err
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return models.TimeEntry{}, fmt.Errorf("invalid date, expected YYYY-MM-DD")
	}
	te := models.TimeEntry{
		AssignmentID:    req.AssignmentID,
		Date:            date,
		Hours:           req.Hours,
		TaskDescription: req.TaskDescription,
	}
	if req.BillingCodeID != nil {
		te.BillingCodeID = *req.BillingCodeID
	}
	return te, nil
}

func (s *WebServer) handleGenerateInvoice(c *gin.Context) {
	pID, err := strconv.Atoi(c.Query("proj"))
	if err != nil || pID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project"})
		return
	}
	eID, err := strconv.Atoi(c.Query("emp"))
	if err != nil || eID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee"})
		return
	}
	yr, err := strconv.Atoi(c.Query("year"))
	if err != nil || yr < 2000 || yr > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	mVal, err := strconv.Atoi(c.Query("month"))
	if err != nil || mVal < 1 || mVal > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}
	descriptionMode := c.DefaultQuery("description_mode", "task")
	if descriptionMode != "task" && descriptionMode != "project" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid description mode"})
		return
	}

	pdfData, err := s.invSvc.GenerateMonthlyInvoicePDF(pID, eID, yr, time.Month(mVal), descriptionMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileName := fmt.Sprintf("invoice_%d_%d_%04d_%02d.pdf", pID, eID, yr, mVal)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

func (s *WebServer) Start(port string) error {
	return s.router.Run(":" + port)
}
