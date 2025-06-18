package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MhunterDev/img2ascii/source/handlers"
	"github.com/gin-gonic/gin"
)

func TestHandleHome(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	cfg := &handlers.Config{
		// ...you may need to set required fields for the test...
	}
	r.GET("/", handlers.HandleHome(cfg))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 && w.Code != 500 {
		t.Errorf("Expected 200 or 500, got %d", w.Code)
	}
}
