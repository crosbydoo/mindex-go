package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuccessEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ok", func(c *gin.Context) {
		OK(c, "Entries retrieved successfully", []string{"a"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body.Code != 200 || !body.Status {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body.Message != "Entries retrieved successfully" {
		t.Fatalf("unexpected message: %q", body.Message)
	}
	if body.Data == nil {
		t.Fatal("expected data")
	}
}

func TestErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/err", func(c *gin.Context) {
		Unauthorized(c, "Unauthorized")
	})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var body Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body.Code != 401 || body.Status {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body.Data != nil {
		t.Fatalf("expected nil data, got %#v", body.Data)
	}
}
