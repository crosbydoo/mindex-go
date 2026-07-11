package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"mindex-api/core/domain"
	"mindex-api/core/repository"
	"mindex-api/core/service"
	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func setupEntryRouter() (*gin.Engine, service.EntryService) {
	gin.SetMode(gin.TestMode)
	repo := repository.NewEntryRepositoryMock([]domain.Entry{
		{ID: 1, Title: "Entry 1", Abstract: "A", Category: "Clinical Psychology", Year: 2024, Author: "A", Source: "S", Type: "Journal", URL: "#"},
		{ID: 2, Title: "Entry 2", Abstract: "B", Category: "Mental Health", Year: 2023, Author: "B", Source: "S", Type: "Article", URL: "#"},
	})
	categoryRepo := repository.NewCategoryRepositoryMock(domain.CategoryList, repo)
	svc := service.NewEntryService(repo, categoryRepo)
	handler := NewEntryHandler(svc)
	categoryHandler := NewCategoryHandler(service.NewCategoryService(categoryRepo))

	r := gin.New()
	r.GET("/api/entries", handler.List)
	r.GET("/api/categories", handler.ListByCategories)
	r.GET("/api/categories/list", categoryHandler.List)
	r.POST("/api/categories", categoryHandler.Create)
	r.PUT("/api/categories", categoryHandler.Update)
	r.DELETE("/api/categories", categoryHandler.Delete)
	r.POST("/api/entries", handler.Create)
	r.PUT("/api/entries", handler.Update)
	r.DELETE("/api/entries", handler.Delete)
	r.POST("/api/entries/archive", handler.Archive)
	r.POST("/api/entries/unarchive", handler.Unarchive)

	return r, svc
}

func TestEntryHandler_List_Paginated(t *testing.T) {
	r, _ := setupEntryRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/entries?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status true")
	}

	raw, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	var data domain.PaginatedEntries
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("failed to decode data: %v", err)
	}
	if len(data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(data.Items))
	}
	if data.Pagination.Total != 2 {
		t.Fatalf("expected total 2, got %d", data.Pagination.Total)
	}
}

func TestEntryHandler_ListByCategories(t *testing.T) {
	r, _ := setupEntryRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/categories?page=1&limit=5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status true")
	}

	raw, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	var data domain.CategoriesResult
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("failed to decode data: %v", err)
	}
	if len(data.Categories) != len(domain.CategoryList) {
		t.Fatalf("expected %d categories, got %d", len(domain.CategoryList), len(data.Categories))
	}
}

func TestEntryHandler_Create_UnauthorizedPayload(t *testing.T) {
	r, _ := setupEntryRouter()

	body := bytes.NewBufferString(`{"title":"Only title"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/entries", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status {
		t.Fatal("expected status false")
	}
}

func TestEntryHandler_ArchiveAndListArchived(t *testing.T) {
	r, _ := setupEntryRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/entries/archive?id=1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	raw, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	var entry domain.Entry
	if err := json.Unmarshal(raw, &entry); err != nil {
		t.Fatalf("failed to decode entry: %v", err)
	}
	if !entry.IsArchived {
		t.Fatal("expected is_archived true")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/entries?archived=true", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRec.Code)
	}

	var listResp response.Body
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	listRaw, err := json.Marshal(listResp.Data)
	if err != nil {
		t.Fatalf("failed to marshal list data: %v", err)
	}
	var data domain.PaginatedEntries
	if err := json.Unmarshal(listRaw, &data); err != nil {
		t.Fatalf("failed to decode list data: %v", err)
	}
	if len(data.Items) != 1 {
		t.Fatalf("expected 1 archived item, got %d", len(data.Items))
	}
}

func TestCategoryHandler_CreateListDelete(t *testing.T) {
	r, _ := setupEntryRouter()

	body := bytes.NewBufferString(`{"name":"New Psychology Field"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/categories", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createResp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	raw, _ := json.Marshal(createResp.Data)
	var created domain.Category
	if err := json.Unmarshal(raw, &created); err != nil {
		t.Fatalf("failed to decode category: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/categories/list", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRec.Code)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/categories?id="+strconv.FormatInt(created.ID, 10), nil)
	delRec := httptest.NewRecorder()
	r.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", delRec.Code, delRec.Body.String())
	}
}

func TestCategoryHandler_DeleteInUse(t *testing.T) {
	r, _ := setupEntryRouter()

	req := httptest.NewRequest(http.MethodDelete, "/api/categories?id=1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLoginHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginHandler := NewLoginHandler(service.NewLoginService("secret"))

	r := gin.New()
	r.POST("/api/login", loginHandler.Login)

	body := bytes.NewBufferString(`{"password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status true")
	}

	raw, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("failed to decode data: %v", err)
	}
	if data.Token == "" {
		t.Fatal("expected token in response data")
	}
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginHandler := NewLoginHandler(service.NewLoginService("secret"))

	r := gin.New()
	r.POST("/api/login", loginHandler.Login)

	body := bytes.NewBufferString(`{"password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status {
		t.Fatal("expected status false")
	}
}

func TestLoginHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginHandler := NewLoginHandler(service.NewLoginService("secret"))

	r := gin.New()
	r.POST("/api/logout", loginHandler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status true")
	}
}

func TestLoginHandler_AdminNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loginHandler := NewLoginHandler(service.NewLoginService(""))

	r := gin.New()
	r.POST("/api/login", loginHandler.Login)

	body := bytes.NewBufferString(`{"password":"anything"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
