package web

import (
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
		api.GET("/projects", s.handleGetProjects)
		api.POST("/projects", s.handleCreateProject)
		api.GET("/assignments", s.handleGetAssignments)
		api.POST("/assignments", s.handleCreateAssignment)
		api.GET("/time", s.handleGetTimeEntries)
		api.POST("/time", s.handleCreateTimeEntry)
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
	rows, err := s.timeSvc.GetDB().Query("SELECT id, assignment_id, date, hours, task_description FROM time_entries")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []map[string]interface{}
	for rows.Next() {
		var id, asnID int
		var date string
		var hrs float64
		var desc string
		if err := rows.Scan(&id, &asnID, &date, &hrs, &desc); err != nil {
			continue
		}
		entries = append(entries, map[string]interface{}{
			"id": id, "assignment_id": asnID, "date": date, "hours": hrs, "task_description": desc,
		})
	}
	c.JSON(http.StatusOK, entries)
}

func (s *WebServer) handleCreateTimeEntry(c *gin.Context) {
	var te models.TimeEntry
	if err := c.ShouldBindJSON(&te); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.timeSvc.CreateTimeEntry(&te); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, te)
}

func (s *WebServer) handleGenerateInvoice(c *gin.Context) {
	pID, _ := strconv.Atoi(c.Query("proj"))
	eID, _ := strconv.Atoi(c.Query("emp"))
	yr, _ := strconv.Atoi(c.Query("year"))
	mVal, _ := strconv.Atoi(c.Query("month"))

	fileName := fmt.Sprintf("invoice_%d_%d_%d.pdf", pID, eID, mVal)

	if err := s.invSvc.GenerateMonthlyInvoice(pID, eID, yr, time.Month(mVal), fileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(fileName)
}

func (s *WebServer) Start(port string) error {
	return s.router.Run(":" + port)
}
