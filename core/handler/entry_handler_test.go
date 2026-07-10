package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mindex-api/core/domain"
	"mindex-api/core/repository"
	"mindex-api/core/service"

	"github.com/gin-gonic/gin"
)

func setupEntryRouter() (*gin.Engine, service.EntryService) {
	gin.SetMode(gin.TestMode)
	repo := repository.NewEntryRepositoryMock(nil)
	svc := service.NewEntryService(repo)
	handler := NewEntryHandler(svc)

	r := gin.New()
	r.GET("/api/entries", handler.List)
	r.POST("/api/entries", handler.Create)
	r.PUT("/api/entries", handler.Update)
	r.DELETE("/api/entries", handler.Delete)

	return r, svc
}

func TestEntryHandler_List_Empty(t *testing.T) {
	r, _ := setupEntryRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []domain.Entry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(entries))
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

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token in response")
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

	var resp struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.OK {
		t.Fatal("expected ok=true in response")
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
