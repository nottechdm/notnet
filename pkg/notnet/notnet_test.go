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

	wantRead := 10 * time.Second
	wantWrite := 20 * time.Second

	engine.GET("/custom", func(req *Request, res *Response) error {
		return res.String(200, "ok")
	}).ApplyConfig(&EngineOption{
		ReadTimeout:  wantRead,
		WriteTimeout: wantWrite,
	})

	// Verify the engine-wide timeouts were updated.
	if engine.readTimeout != wantRead {
		t.Errorf("expected readTimeout %v, got %v", wantRead, engine.readTimeout)
	}
	if engine.writeTimeout != wantWrite {
		t.Errorf("expected writeTimeout %v, got %v", wantWrite, engine.writeTimeout)
	}

	// Also verify the route still responds correctly.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/custom", nil)
	engine.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestApplyConfig_UpdatesEngineFields(t *testing.T) {
	engine := New(nil)

	wantRead := 5 * time.Second
	wantWrite := 8 * time.Second
	wantIdle := 30 * time.Second
	wantMaxHeader := 2048

	engine.ApplyConfig(&EngineOption{
		ReadTimeout:    wantRead,
		WriteTimeout:   wantWrite,
		IdleTimeout:    wantIdle,
		MaxHeaderBytes: wantMaxHeader,
	})

	if engine.readTimeout != wantRead {
		t.Errorf("expected readTimeout %v, got %v", wantRead, engine.readTimeout)
	}
	if engine.writeTimeout != wantWrite {
		t.Errorf("expected writeTimeout %v, got %v", wantWrite, engine.writeTimeout)
	}
	if engine.idleTimeout != wantIdle {
		t.Errorf("expected idleTimeout %v, got %v", wantIdle, engine.idleTimeout)
	}
	if engine.maxHeaderBytes != wantMaxHeader {
		t.Errorf("expected maxHeaderBytes %d, got %d", wantMaxHeader, engine.maxHeaderBytes)
	}
}

func TestApplyConfig_NilIsNoOp(t *testing.T) {
	engine := New(nil)
	before := engine.readTimeout

	engine.ApplyConfig(nil)

	if engine.readTimeout != before {
		t.Errorf("expected readTimeout to remain %v after nil ApplyConfig, got %v", before, engine.readTimeout)
	}
}

func TestRouteBuilderMethods(t *testing.T) {
	engine := New(nil)
	rb := engine.GET("/get", func(req *Request, res *Response) error { return nil })

	// Test RouteBuilder chaining and HTTP methods
	rb.POST("/post", func(req *Request, res *Response) error { return nil }).
		PUT("/put", func(req *Request, res *Response) error { return nil }).
		DELETE("/delete", func(req *Request, res *Response) error { return nil }).
		PATCH("/patch", func(req *Request, res *Response) error { return nil }).
		OPTIONS("/options", func(req *Request, res *Response) error { return nil }).
		HEAD("/head", func(req *Request, res *Response) error { return nil }).
		Use(func(req *Request, res *Response) error { return req.Next() }).
		SetErrorHandler(func(req *Request, res *Response, err error) {}).
		SetNotFoundHandler(func(req *Request, res *Response) {}).
		SetPanicHandler(func(req *Request, res *Response, rec interface{}) {})

	// Verify routes are registered
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	paths := []string{"/get", "/post", "/put", "/delete", "/patch", "/options", "/head"}

	for i, method := range methods {
		_, _, found := engine.router.Match(method, paths[i])
		if !found {
			t.Errorf("expected route %s %s to be registered", method, paths[i])
		}
	}
}

func TestEngineGetters(t *testing.T) {
	opts := &EngineOption{
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   20 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 4096,
	}
	engine := New(opts)

	if engine.GetReadTimeout() != opts.ReadTimeout {
		t.Errorf("expected %v, got %v", opts.ReadTimeout, engine.GetReadTimeout())
	}
	if engine.GetWriteTimeout() != opts.WriteTimeout {
		t.Errorf("expected %v, got %v", opts.WriteTimeout, engine.GetWriteTimeout())
	}
	if engine.GetIdleTimeout() != opts.IdleTimeout {
		t.Errorf("expected %v, got %v", opts.IdleTimeout, engine.GetIdleTimeout())
	}
	if engine.GetMaxHeaderBytes() != opts.MaxHeaderBytes {
		t.Errorf("expected %d, got %d", opts.MaxHeaderBytes, engine.GetMaxHeaderBytes())
	}
}

func TestRouteBuilderApplyConfig(t *testing.T) {
	engine := New(nil)
	rb := engine.GET("/test", func(req *Request, res *Response) error { return nil })
	
	rb.ApplyConfig(&EngineOption{
		ReadTimeout: 12 * time.Second,
	})

	if engine.readTimeout != 12*time.Second {
		t.Errorf("expected engine readTimeout 12s, got %v", engine.readTimeout)
	}
}

func TestGroupApplyConfig(t *testing.T) {
	engine := New(nil)
	group := engine.Group("/api")
	
	group.ApplyConfig(&EngineOption{
		WriteTimeout: 14 * time.Second,
	})

	if engine.writeTimeout != 14*time.Second {
		t.Errorf("expected engine writeTimeout 14s, got %v", engine.writeTimeout)
	}
}

func TestRouteBuilderGroup(t *testing.T) {
	engine := New(nil)
	rb := engine.GET("/test", func(req *Request, res *Response) error { return nil })
	
	group := rb.Group("/v1")
	if group.prefix != "/v1" {
		t.Errorf("expected group prefix /v1, got %s", group.prefix)
	}
}

func TestListenListener(t *testing.T) {
	engine := New(nil)
	engine.GET("/ping", func(req *Request, res *Response) error {
		return res.String(200, "pong")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	
	go func() {
		_ = engine.ListenListener(ln)
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + ln.Addr().String() + "/ping")
	if err != nil {
		t.Fatalf("failed to GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	
	_ = engine.Shutdown()
}

func TestRouteBuilderSetHandlers(t *testing.T) {
	engine := New(nil)
	rb := engine.GET("/test", func(req *Request, res *Response) error { return nil })
	
	rb.SetNotFoundHandler(func(req *Request, res *Response) {})
	rb.SetPanicHandler(func(req *Request, res *Response, rec interface{}) {})
}

func TestListenError(t *testing.T) {
	engine := New(nil)
	// Invalid address
	err := engine.Listen("invalid")
	if err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestListenTLSError(t *testing.T) {
	engine := New(nil)
	// Invalid certs
	err := engine.ListenTLS(":0", "nonexistent", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent certs")
	}
}
