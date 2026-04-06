package notnet

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	engine := New(nil)
	if engine == nil {
		t.Fatal("expected engine to be created")
	}
	if engine.maxHeaderBytes != 1<<20 {
		t.Error("expected default maxHeaderBytes to be 1MB")
	}
	if engine.readTimeout != 15*time.Second {
		t.Error("expected default readTimeout to be 15s")
	}
}

func TestNewEngineWithOptions(t *testing.T) {
	opts := &EngineOption{
		MaxHeaderBytes: 2048,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
	}
	engine := New(opts)
	if engine.maxHeaderBytes != 2048 {
		t.Error("expected custom maxHeaderBytes to be 2048")
	}
}

func TestEngineUse(t *testing.T) {
	e := New(nil)
	e.Use(func(req *Request, res *Response) error {
		return req.Next()
	})
	if len(e.middleware) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(e.middleware))
	}
}

func TestEngineGET(t *testing.T) {
	engine := New(nil)
	engine.GET("/test", func(req *Request, res *Response) error {
		return res.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestEnginePOST(t *testing.T) {
	engine := New(nil)
	engine.POST("/test", func(req *Request, res *Response) error {
		return res.String(201, "created")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	engine.ServeHTTP(w, r)

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestEnginePUT(t *testing.T) {
	engine := New(nil)
	engine.PUT("/test", func(req *Request, res *Response) error {
		return res.String(200, "updated")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/test", strings.NewReader(""))
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestEngineDELETE(t *testing.T) {
	engine := New(nil)
	engine.DELETE("/test", func(req *Request, res *Response) error {
		return res.String(204, "")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/test", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestEnginePATCH(t *testing.T) {
	engine := New(nil)
	engine.PATCH("/test", func(req *Request, res *Response) error {
		return res.String(200, "patched")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PATCH", "/test", strings.NewReader(""))
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestEngineOPTIONS(t *testing.T) {
	engine := New(nil)
	engine.OPTIONS("/test", func(req *Request, res *Response) error {
		return res.String(204, "")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/test", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestEngineHEAD(t *testing.T) {
	engine := New(nil)
	engine.HEAD("/test", func(req *Request, res *Response) error {
		return res.String(200, "")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("HEAD", "/test", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestEngineGroup(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api/v1", AuthRequired())

	if group.prefix != "/api/v1" {
		t.Errorf("expected prefix /api/v1, got %s", group.prefix)
	}
	if group.engine != engine {
		t.Error("expected group.engine to be the same as engine")
	}
}

func TestGroupUse(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api/v1")
	group.Use(Logger())

	if len(group.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(group.handlers))
	}
}

func TestGroupGET(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api")
	group.GET("/test", func(req *Request, res *Response) error {
		return res.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGroupPOST(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api")
	group.POST("/test", func(req *Request, res *Response) error {
		return res.String(201, "created")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/test", strings.NewReader(""))
	engine.ServeHTTP(w, r)

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestEngineSetErrorHandler(t *testing.T) {
	engine := New(nil)
	engine.SetErrorHandler(func(req *Request, res *Response, err error) {})

	if engine.errorFunc == nil {
		t.Error("expected errorFunc to be set")
	}
}

func TestEngineSetNotFoundHandler(t *testing.T) {
	engine := New(nil)
	engine.SetNotFoundHandler(func(req *Request, res *Response) {})

	if engine.notFoundFunc == nil {
		t.Error("expected notFoundFunc to be set")
	}
}

func TestEngineSetPanicHandler(t *testing.T) {
	engine := New(nil)
	engine.SetPanicHandler(func(req *Request, res *Response, rec interface{}) {})

	if engine.panicFunc == nil {
		t.Error("expected panicFunc to be set")
	}
}

func TestServeHTTPNotFound(t *testing.T) {
	engine := New(nil)
	engine.GET("/exists", func(req *Request, res *Response) error {
		return res.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/notfound", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestServeHTTPWithMiddleware(t *testing.T) {
	engine := New(nil)

	middlewareCalled := false
	engine.Use(func(req *Request, res *Response) error {
		middlewareCalled = true
		return req.Next()
	})

	engine.GET("/test", func(req *Request, res *Response) error {
		return res.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, r)

	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
}

func TestServeHTTPError(t *testing.T) {
	engine := New(nil)
	errorCalled := false

	engine.SetErrorHandler(func(req *Request, res *Response, err error) {
		errorCalled = true
		res.JSON(500, map[string]string{"error": "custom"})
	})

	engine.GET("/test", func(req *Request, res *Response) error {
		return fmt.Errorf("test error")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, r)

	if !errorCalled {
		t.Error("expected error handler to be called")
	}
}

func TestServeHTTPPanic(t *testing.T) {
	engine := New(nil)
	panicCalled := false

	engine.SetPanicHandler(func(req *Request, res *Response, rec interface{}) {
		panicCalled = true
		res.JSON(500, map[string]string{"error": "panic recovered"})
	})

	engine.GET("/test", func(req *Request, res *Response) error {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, r)

	if !panicCalled {
		t.Error("expected panic handler to be called")
	}
}

func TestEngineShutdown(t *testing.T) {
	engine := New(nil)
	err := engine.Shutdown()
	if err != nil {
		t.Error("expected no error when shutting down without server")
	}
}

func TestEngineChaining(t *testing.T) {
	engine := New(nil)

	result := engine.
		Use(Logger()).
		GET("/1", func(req *Request, res *Response) error { return nil }).
		POST("/2", func(req *Request, res *Response) error { return nil }).
		PUT("/3", func(req *Request, res *Response) error { return nil }).
		DELETE("/4", func(req *Request, res *Response) error { return nil }).
		SetErrorHandler(func(req *Request, res *Response, err error) {}).
		SetNotFoundHandler(func(req *Request, res *Response) {}).
		SetPanicHandler(func(req *Request, res *Response, rec interface{}) {})

	if result != engine {
		t.Error("expected methods to return engine for chaining")
	}
}

func TestGroupChaining(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api")

	result := group.
		Use(Logger()).
		GET("/1", func(req *Request, res *Response) error { return nil }).
		POST("/2", func(req *Request, res *Response) error { return nil }).
		PUT("/3", func(req *Request, res *Response) error { return nil }).
		DELETE("/4", func(req *Request, res *Response) error { return nil }).
		PATCH("/5", func(req *Request, res *Response) error { return nil })

	if result != group {
		t.Error("expected methods to return group for chaining")
	}
}

func TestGroupWrapHandler(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api")

	handlerCalled := false
	handler := func(req *Request, res *Response) error {
		handlerCalled = true
		return res.String(200, "ok")
	}

	wrappedHandler := group.wrapHandler(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	wrappedHandler(req, res)

	if !handlerCalled {
		t.Error("expected wrapped handler to be called")
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	engine := New(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	engine.defaultErrorHandler(req, res, fmt.Errorf("test error"))

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestDefaultNotFoundHandler(t *testing.T) {
	engine := New(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	engine.defaultNotFoundHandler(req, res)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDefaultPanicHandler(t *testing.T) {
	engine := New(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	engine.defaultPanicHandler(req, res, "test panic")

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestApplyConfig(t *testing.T) {
	engine := New(nil)

	// Apply custom config to a route
	engine.GET("/custom", func(req *Request, res *Response) error {
		return res.String(200, "ok")
	}).ApplyConfig(&EngineOption{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	})

	// Verify that the route works
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/custom", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestApplyConfig_Fail(t *testing.T) {
	engine := New(nil)

	// Apply custom config to a route
	engine.GET("/custom", func(req *Request, res *Response) error {
		time.Sleep(50 * time.Millisecond) // Simulate a long processing time
		return res.String(200, "ok")
	}).ApplyConfig(&EngineOption{
		ReadTimeout:  1 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	go engine.ListenListener(ln)
	defer engine.Shutdown()

	time.Sleep(10 * time.Millisecond) // Server startup wait

	client := &http.Client{Timeout: 500 * time.Millisecond}
	url := "http://" + ln.Addr().String() + "/custom"
	resp, err := client.Get(url)
	if err == nil {
		defer resp.Body.Close()
		t.Errorf("expected request to fail due to timeout, but got status %d", resp.StatusCode)
	}
}
