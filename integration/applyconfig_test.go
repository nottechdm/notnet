package integration

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nottechdm/notnet/pkg/notnet"
)

// TestApplyConfigAfterGetRoute tests applying config after registering a GET route
func TestApplyConfigAfterGetRoute(t *testing.T) {
	app := notnet.New(nil)

	wantRead := 20 * time.Second
	wantWrite := 40 * time.Second

	// Apply config after registering a GET route
	app.GET("/fast-api", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "ok")
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout:  wantRead,
		WriteTimeout: wantWrite,
	})

	// Verify engine-wide timeouts were updated
	if app.GetReadTimeout() != wantRead {
		t.Errorf("expected readTimeout %v, got %v", wantRead, app.GetReadTimeout())
	}
	if app.GetWriteTimeout() != wantWrite {
		t.Errorf("expected writeTimeout %v, got %v", wantWrite, app.GetWriteTimeout())
	}

	// Verify route works correctly
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/fast-api", nil)
	app.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", w.Body.String())
	}
}

// TestApplyConfigAfterPostRoute tests applying config after registering a POST route
func TestApplyConfigAfterPostRoute(t *testing.T) {
	app := notnet.New(nil)

	wantRead := 15 * time.Second

	app.POST("/users", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(201, "created")
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout: wantRead,
	})

	if app.GetReadTimeout() != wantRead {
		t.Errorf("expected readTimeout %v, got %v", wantRead, app.GetReadTimeout())
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/users", strings.NewReader(""))
	app.ServeHTTP(w, r)

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
	if w.Body.String() != "created" {
		t.Errorf("expected body 'created', got %s", w.Body.String())
	}
}

// TestApplyConfigChainedMultipleRoutes tests chaining multiple route registrations with ApplyConfig
func TestApplyConfigChainedMultipleRoutes(t *testing.T) {
	app := notnet.New(nil)

	// Register multiple routes with chained configurations
	app.POST("/users", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(201, "created")
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout: 5 * time.Second,
	}).PUT("/users/:id", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "updated")
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout: 8 * time.Second,
	}).DELETE("/users/:id", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(204, "")
	})

	// Verify final configuration
	if app.GetReadTimeout() != 8*time.Second {
		t.Errorf("expected readTimeout 8s, got %v", app.GetReadTimeout())
	}

	// Verify all routes work correctly
	testCases := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
		wantBody string
	}{
		{"POST users", "POST", "/users", "", 201, "created"},
		{"PUT user", "PUT", "/users/1", "", 200, "updated"},
		{"DELETE user", "DELETE", "/users/1", "", 204, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			app.ServeHTTP(w, r)

			if w.Code != tc.wantCode {
				t.Errorf("expected status %d, got %d", tc.wantCode, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Errorf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

// TestApplyConfigWithMiddleware tests ApplyConfig with middleware
func TestApplyConfigWithMiddleware(t *testing.T) {
	app := notnet.New(nil)

	middlewareCalled := false
	app.Use(func(req *notnet.Request, res *notnet.Response) error {
		middlewareCalled = true
		return req.Next()
	})

	app.GET("/test", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "success")
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout: 10 * time.Second,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	app.ServeHTTP(w, r)

	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", w.Body.String())
	}
}

// TestApplyConfigWithErrorHandler tests ApplyConfig with error handling
func TestApplyConfigWithErrorHandler(t *testing.T) {
	app := notnet.New(nil)

	errorHandled := false
	app.GET("/error", func(req *notnet.Request, res *notnet.Response) error {
		return &testError{message: "test error"}
	}).SetErrorHandler(func(req *notnet.Request, res *notnet.Response, err error) {
		errorHandled = true
		res.JSON(500, map[string]string{"error": "handled"})
	}).ApplyConfig(&notnet.EngineOption{
		WriteTimeout: 15 * time.Second,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/error", nil)
	app.ServeHTTP(w, r)

	if !errorHandled {
		t.Error("expected error handler to be called")
	}
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// TestApplyConfigWithGroup tests ApplyConfig with route groups
func TestApplyConfigWithGroup(t *testing.T) {
	app := notnet.New(nil)

	api := app.Group("/api")

	api.GET("/health", func(req *notnet.Request, res *notnet.Response) error {
		return res.JSON(200, map[string]string{"status": "ok"})
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout: 12 * time.Second,
	})

	if app.GetReadTimeout() != 12*time.Second {
		t.Errorf("expected readTimeout 12s, got %v", app.GetReadTimeout())
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/health", nil)
	app.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestApplyConfigMaxHeaderBytes tests setting MaxHeaderBytes via ApplyConfig
func TestApplyConfigMaxHeaderBytes(t *testing.T) {
	app := notnet.New(nil)

	wantMaxHeader := 4096

	app.GET("/upload", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "ok")
	}).ApplyConfig(&notnet.EngineOption{
		MaxHeaderBytes: wantMaxHeader,
	})

	if app.GetMaxHeaderBytes() != wantMaxHeader {
		t.Errorf("expected maxHeaderBytes %d, got %d", wantMaxHeader, app.GetMaxHeaderBytes())
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/upload", nil)
	app.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
