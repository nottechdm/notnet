package notnet

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestDashboard(t *testing.T) {
	engine := New(nil)
	engine.GET("/dashboard", Dashboard())

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/dashboard", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content-type text/html; charset=utf-8, got %s", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestStatsAPI(t *testing.T) {
	engine := New(nil)
	engine.GET("/api/stats", StatsAPI())

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/stats", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected content-type application/json; charset=utf-8, got %s", contentType)
	}

	var stats StatsData
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	if err != nil {
		t.Errorf("failed to unmarshal stats: %v", err)
	}

	if stats.UpTime <= 0 {
		t.Errorf("expected positive uptime, got %v", stats.UpTime)
	}
}
