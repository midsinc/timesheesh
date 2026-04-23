package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"timesheesh/internal/db"
	"timesheesh/internal/services"
)

func newTestServer(t *testing.T) *WebServer {
	t.Helper()

	database, err := db.InitDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	timeSvc := services.NewTimeService(database)
	invSvc := services.NewInvoiceService(database)
	return NewWebServer(timeSvc, invSvc)
}

func TestCreateBillingCodeRouteExists(t *testing.T) {
	server := newTestServer(t)

	projectReq := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{"name":"Alpha","company_name":"Acme","description":"Test project"}`))
	projectReq.Header.Set("Content-Type", "application/json")
	projectRec := httptest.NewRecorder()
	server.router.ServeHTTP(projectRec, projectReq)

	if projectRec.Code != http.StatusOK {
		t.Fatalf("expected project create status 200, got %d", projectRec.Code)
	}

	billingReq := httptest.NewRequest(http.MethodPost, "/api/billing-codes", bytes.NewBufferString(`{"project_id":1,"code":"DEV-01","description":"Development"}`))
	billingReq.Header.Set("Content-Type", "application/json")
	billingRec := httptest.NewRecorder()
	server.router.ServeHTTP(billingRec, billingReq)

	if billingRec.Code != http.StatusOK {
		t.Fatalf("expected billing code create status 200, got %d with body %s", billingRec.Code, billingRec.Body.String())
	}
}

func TestCreateTimeEntryWithBillingCodeAndDateString(t *testing.T) {
	server := newTestServer(t)

	requests := []struct {
		path string
		body string
	}{
		{path: "/api/employees", body: `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`},
		{path: "/api/projects", body: `{"name":"Alpha","company_name":"Acme","description":"Test project"}`},
		{path: "/api/assignments", body: `{"employee_id":1,"project_id":1,"billable_rate":150,"pay_rate":90}`},
		{path: "/api/billing-codes", body: `{"project_id":1,"code":"DEV-01","description":"Development"}`},
	}

	for _, reqSpec := range requests {
		req := httptest.NewRequest(http.MethodPost, reqSpec.path, bytes.NewBufferString(reqSpec.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d with body %s", reqSpec.path, rec.Code, rec.Body.String())
		}
	}

	timeReq := httptest.NewRequest(http.MethodPost, "/api/time", bytes.NewBufferString(`{"assignment_id":1,"billing_code_id":1,"date":"2026-04-12","hours":8,"task_description":"Build feature"}`))
	timeReq.Header.Set("Content-Type", "application/json")
	timeRec := httptest.NewRecorder()
	server.router.ServeHTTP(timeRec, timeReq)

	if timeRec.Code != http.StatusOK {
		t.Fatalf("expected time entry create status 200, got %d with body %s", timeRec.Code, timeRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/time?month=2026-04", nil)
	listRec := httptest.NewRecorder()
	server.router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected time entry list status 200, got %d with body %s", listRec.Code, listRec.Body.String())
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte(`"billing_code_id":1`)) {
		t.Fatalf("expected billing_code_id in time entry list, got %s", listRec.Body.String())
	}
}

func TestUpdateTimeEntry(t *testing.T) {
	server := newTestServer(t)

	requests := []struct {
		path string
		body string
	}{
		{path: "/api/employees", body: `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`},
		{path: "/api/projects", body: `{"name":"Alpha","company_name":"Acme","description":"Test project"}`},
		{path: "/api/assignments", body: `{"employee_id":1,"project_id":1,"billable_rate":150,"pay_rate":90}`},
		{path: "/api/billing-codes", body: `{"project_id":1,"code":"DEV-01","description":"Development"}`},
		{path: "/api/time", body: `{"assignment_id":1,"date":"2026-04-12","hours":8,"task_description":"Initial entry"}`},
	}

	for _, reqSpec := range requests {
		req := httptest.NewRequest(http.MethodPost, reqSpec.path, bytes.NewBufferString(reqSpec.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d with body %s", reqSpec.path, rec.Code, rec.Body.String())
		}
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/time/1", bytes.NewBufferString(`{"assignment_id":1,"billing_code_id":1,"date":"2026-04-15","hours":6.5,"task_description":"Updated entry"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	server.router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d with body %s", updateRec.Code, updateRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/time?month=2026-04", nil)
	listRec := httptest.NewRecorder()
	server.router.ServeHTTP(listRec, listReq)

	body := listRec.Body.Bytes()
	if !bytes.Contains(body, []byte(`"billing_code_id":1`)) || !bytes.Contains(body, []byte(`"date":"2026-04-15"`)) || !bytes.Contains(body, []byte(`"hours":6.5`)) {
		t.Fatalf("expected updated time entry in list, got %s", listRec.Body.String())
	}
}

func TestGenerateInvoiceReturnsPDF(t *testing.T) {
	server := newTestServer(t)

	requests := []struct {
		path string
		body string
	}{
		{path: "/api/employees", body: `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`},
		{path: "/api/projects", body: `{"name":"Alpha","company_name":"Acme","description":"Test project"}`},
		{path: "/api/assignments", body: `{"employee_id":1,"project_id":1,"billable_rate":150,"pay_rate":90}`},
		{path: "/api/time", body: `{"assignment_id":1,"date":"2026-04-12","hours":8,"task_description":"Build feature"}`},
	}

	for _, reqSpec := range requests {
		req := httptest.NewRequest(http.MethodPost, reqSpec.path, bytes.NewBufferString(reqSpec.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d with body %s", reqSpec.path, rec.Code, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/invoice?proj=1&emp=1&year=2026&month=4&description_mode=project", nil)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected invoice status 200, got %d with body %s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/pdf" {
		t.Fatalf("expected application/pdf content type, got %q", contentType)
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) {
		t.Fatalf("expected PDF response, got %q", rec.Body.String())
	}
}

func TestGenerateInvoiceRejectsInvalidDescriptionMode(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/invoice?proj=1&emp=1&year=2026&month=4&description_mode=invalid", nil)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invoice status 400, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestGenerateInvoiceRejectsInvalidMonth(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/invoice?proj=1&emp=1&year=2026&month=13", nil)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invoice status 400, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateEmployeeProjectAndAssignment(t *testing.T) {
	server := newTestServer(t)

	requests := []struct {
		path string
		body string
	}{
		{path: "/api/employees", body: `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","address":"Old Address"}`},
		{path: "/api/projects", body: `{"name":"Alpha","company_name":"Acme","description":"Test project","default_payment_terms":30}`},
		{path: "/api/assignments", body: `{"employee_id":1,"project_id":1,"billable_rate":150,"pay_rate":90}`},
	}

	for _, reqSpec := range requests {
		req := httptest.NewRequest(http.MethodPost, reqSpec.path, bytes.NewBufferString(reqSpec.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d with body %s", reqSpec.path, rec.Code, rec.Body.String())
		}
	}

	updates := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPut, path: "/api/employees/1", body: `{"first_name":"Grace","last_name":"Hopper","email":"grace@example.com","address":"New Address"}`},
		{method: http.MethodPut, path: "/api/projects/1", body: `{"name":"Beta","company_name":"Globex","description":"Updated project","default_payment_terms":45}`},
		{method: http.MethodPut, path: "/api/assignments/1", body: `{"employee_id":1,"project_id":1,"billable_rate":175,"pay_rate":95}`},
	}

	for _, reqSpec := range updates {
		req := httptest.NewRequest(reqSpec.method, reqSpec.path, bytes.NewBufferString(reqSpec.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d with body %s", reqSpec.path, rec.Code, rec.Body.String())
		}
	}

	employeesReq := httptest.NewRequest(http.MethodGet, "/api/employees", nil)
	employeesRec := httptest.NewRecorder()
	server.router.ServeHTTP(employeesRec, employeesReq)
	if !bytes.Contains(employeesRec.Body.Bytes(), []byte(`"first_name":"Grace"`)) || !bytes.Contains(employeesRec.Body.Bytes(), []byte(`"address":"New Address"`)) {
		t.Fatalf("expected updated employee, got %s", employeesRec.Body.String())
	}

	projectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	projectsRec := httptest.NewRecorder()
	server.router.ServeHTTP(projectsRec, projectsReq)
	if !bytes.Contains(projectsRec.Body.Bytes(), []byte(`"name":"Beta"`)) || !bytes.Contains(projectsRec.Body.Bytes(), []byte(`"default_payment_terms":45`)) {
		t.Fatalf("expected updated project, got %s", projectsRec.Body.String())
	}

	assignmentsReq := httptest.NewRequest(http.MethodGet, "/api/assignments", nil)
	assignmentsRec := httptest.NewRecorder()
	server.router.ServeHTTP(assignmentsRec, assignmentsReq)
	if !bytes.Contains(assignmentsRec.Body.Bytes(), []byte(`"billable_rate":175`)) || !bytes.Contains(assignmentsRec.Body.Bytes(), []byte(`"pay_rate":95`)) {
		t.Fatalf("expected updated assignment, got %s", assignmentsRec.Body.String())
	}
}
