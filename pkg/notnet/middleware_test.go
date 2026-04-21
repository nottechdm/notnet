package notnet

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		recovery := Recovery()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{
			func(req *Request, res *Response) error {
				return res.String(200, "ok")
			},
		}

		err := recovery(req, res)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("with panic", func(t *testing.T) {
		recovery := Recovery()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{
			func(req *Request, res *Response) error {
				panic("test panic")
			},
		}

		recovery(req, res)

		if w.Code != 500 {
			t.Errorf("expected status 500 after panic, got %d", w.Code)
		}
	})
}

func TestLoggerMiddleware(t *testing.T) {
	logger := Logger()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	req.index = -1
	req.handlers = []HandlerFunc{
		func(req *Request, res *Response) error {
			time.Sleep(10 * time.Millisecond)
			return res.String(200, "ok")
		},
	}

	err := logger(req, res)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthRequiredMiddleware(t *testing.T) {
	authRequired := AuthRequired()

	t.Run("with valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Authorization", "Bearer token123")
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{
			func(req *Request, res *Response) error {
				return res.String(200, "authorized")
			},
		}

		err := authRequired(req, res)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("without token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{}

		err := authRequired(req, res)
		if err == nil {
			t.Error("expected error for missing token")
		}
		if w.Code != 401 {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		config         *CORSConfig
		method         string
		validateHeader func(*httptest.ResponseRecorder)
	}{
		{
			name:   "default config",
			config: nil,
			method: "GET",
			validateHeader: func(w *httptest.ResponseRecorder) {
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Error("expected Allow-Origin header")
				}
				if w.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("expected Allow-Methods header")
				}
				if w.Header().Get("Access-Control-Allow-Headers") == "" {
					t.Error("expected Allow-Headers header")
				}
			},
		},
		{
			name: "custom config",
			config: &CORSConfig{
				AllowOrigins: []string{"https://example.com"},
				AllowMethods: []string{"GET", "POST"},
				AllowHeaders: []string{"Content-Type"},
			},
			method: "GET",
			validateHeader: func(w *httptest.ResponseRecorder) {
				if !strings.Contains(w.Header().Get("Access-Control-Allow-Origin"), "example.com") {
					t.Error("expected custom origin")
				}
				if !strings.Contains(w.Header().Get("Access-Control-Allow-Methods"), "GET") {
					t.Error("expected custom methods")
				}
			},
		},
		{
			name:   "OPTIONS preflight",
			config: nil,
			method: "OPTIONS",
			validateHeader: func(w *httptest.ResponseRecorder) {
				if w.Code != 204 {
					t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cors := CORS(tt.config)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/test", nil)
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			req.index = -1
			req.handlers = []HandlerFunc{}

			cors(req, res)
			tt.validateHeader(w)
		})
	}
}

func TestCORSCustomHeaders(t *testing.T) {
	config := &CORSConfig{
		AllowOrigins: []string{"*"},
		CustomHeaders: map[string]string{
			"X-Custom-Header": "custom-value",
			"X-Another":       "another-value",
		},
	}

	cors := CORS(config)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	req.index = -1
	req.handlers = []HandlerFunc{}

	cors(req, res)

	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Error("expected custom header to be set")
	}
	if w.Header().Get("X-Another") != "another-value" {
		t.Error("expected another header to be set")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("within limit", func(t *testing.T) {
		rateLimiter := RateLimit(3, 1*time.Second)

		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.RemoteAddr = "127.0.0.1:8000"
			req, res := AcquireRequestResponse(w, r)

			req.index = -1
			req.handlers = []HandlerFunc{
				func(req *Request, res *Response) error {
					return res.String(200, "ok")
				},
			}

			err := rateLimiter(req, res)
			ReleaseRequestResponse(req, res)

			if err != nil {
				t.Errorf("request %d: expected no error, got %v", i+1, err)
			}
			if w.Code != 200 {
				t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
			}
		}
	})

	t.Run("exceeds limit", func(t *testing.T) {
		rateLimiter := RateLimit(2, 1*time.Second)

		// First two requests should succeed
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.RemoteAddr = "127.0.0.1:8000"
			req, res := AcquireRequestResponse(w, r)

			req.index = -1
			req.handlers = []HandlerFunc{
				func(req *Request, res *Response) error {
					return res.String(200, "ok")
				},
			}

			rateLimiter(req, res)
			ReleaseRequestResponse(req, res)
		}

		// Third request should be rejected
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "127.0.0.1:8000"
		req, res := AcquireRequestResponse(w, r)

		req.index = -1
		req.handlers = []HandlerFunc{}

		rateLimiter(req, res)
		ReleaseRequestResponse(req, res)

		if w.Code != 429 {
			t.Errorf("expected status 429, got %d", w.Code)
		}
	})

	t.Run("different clients", func(t *testing.T) {
		rateLimiter := RateLimit(1, 1*time.Second)

		// First client makes one request
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest("GET", "/test", nil)
		r1.RemoteAddr = "192.168.1.1:8000"
		req1, res1 := AcquireRequestResponse(w1, r1)
		req1.index = -1
		req1.handlers = []HandlerFunc{func(req *Request, res *Response) error { return res.String(200, "ok") }}
		rateLimiter(req1, res1)
		ReleaseRequestResponse(req1, res1)

		// Second client should also be able to make one request
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", "/test", nil)
		r2.RemoteAddr = "192.168.1.2:8000"
		req2, res2 := AcquireRequestResponse(w2, r2)
		req2.index = -1
		req2.handlers = []HandlerFunc{func(req *Request, res *Response) error { return res.String(200, "ok") }}
		rateLimiter(req2, res2)
		ReleaseRequestResponse(req2, res2)

		if w1.Code != 200 || w2.Code != 200 {
			t.Error("expected both clients to succeed with their first request")
		}
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("with existing header", func(t *testing.T) {
		requestID := RequestID("X-Request-ID")
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Request-ID", "custom-id-123")
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{}

		requestID(req, res)

		if w.Header().Get("X-Request-ID") != "custom-id-123" {
			t.Error("expected custom request ID to be preserved")
		}
	})

	t.Run("without existing header", func(t *testing.T) {
		requestID := RequestID("X-Request-ID")
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{}

		requestID(req, res)

		if w.Header().Get("X-Request-ID") == "" {
			t.Error("expected request ID to be generated")
		}
	})

	t.Run("different header name", func(t *testing.T) {
		requestID := RequestID("X-Trace-ID")
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Trace-ID", "trace-456")
		req, res := AcquireRequestResponse(w, r)
		defer ReleaseRequestResponse(req, res)

		req.index = -1
		req.handlers = []HandlerFunc{}

		requestID(req, res)

		if w.Header().Get("X-Trace-ID") != "trace-456" {
			t.Error("expected trace ID to be set")
		}
	})
}

func TestCORSConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *CORSConfig
		validate func(*CORSConfig)
	}{
		{
			name:   "nil config gets defaults",
			config: nil,
			validate: func(config *CORSConfig) {
				if config == nil {
					t.Error("expected config to be initialized")
				}
			},
		},
		{
			name: "partial config gets filled",
			config: &CORSConfig{
				AllowOrigins: []string{"https://example.com"},
			},
			validate: func(config *CORSConfig) {
				if len(config.AllowOrigins) != 1 {
					t.Error("expected AllowOrigins to be preserved")
				}
				if config.AllowMethods == nil {
					t.Error("expected AllowMethods to be filled")
				}
				if config.AllowHeaders == nil {
					t.Error("expected AllowHeaders to be filled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cors := CORS(tt.config)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			req.index = -1
			req.handlers = []HandlerFunc{}

			cors(req, res)
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	engine := New(nil)

	order := []string{}

	engine.Use(func(req *Request, res *Response) error {
		order = append(order, "middleware1")
		return req.Next()
	})

	engine.Use(func(req *Request, res *Response) error {
		order = append(order, "middleware2")
		return req.Next()
	})

	engine.GET("/test", func(req *Request, res *Response) error {
		order = append(order, "handler")
		return res.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	engine.ServeHTTP(w, r)

	if len(order) != 3 {
		t.Errorf("expected 3 calls, got %d", len(order))
	}
	if order[0] != "middleware1" || order[1] != "middleware2" || order[2] != "handler" {
		t.Errorf("expected correct order, got %v", order)
	}
}

func TestRateLimitWindowReset(t *testing.T) {
	rateLimiter := RateLimit(1, 100*time.Millisecond)

	// First request
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/test", nil)
	r1.RemoteAddr = "127.0.0.1:8000"
	req1, res1 := AcquireRequestResponse(w1, r1)
	req1.index = -1
	req1.handlers = []HandlerFunc{func(req *Request, res *Response) error { return res.String(200, "ok") }}
	rateLimiter(req1, res1)
	ReleaseRequestResponse(req1, res1)

	if w1.Code != 200 {
		t.Error("first request should succeed")
	}

	// Second request immediately after (should fail)
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/test", nil)
	r2.RemoteAddr = "127.0.0.1:8000"
	req2, res2 := AcquireRequestResponse(w2, r2)
	req2.index = -1
	req2.handlers = []HandlerFunc{}
	rateLimiter(req2, res2)
	ReleaseRequestResponse(req2, res2)

	if w2.Code != 429 {
		t.Error("second request should be rate limited")
	}

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Third request after window reset (should succeed)
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/test", nil)
	r3.RemoteAddr = "127.0.0.1:8000"
	req3, res3 := AcquireRequestResponse(w3, r3)
	req3.index = -1
	req3.handlers = []HandlerFunc{func(req *Request, res *Response) error { return res.String(200, "ok") }}
	rateLimiter(req3, res3)
	ReleaseRequestResponse(req3, res3)

	if w3.Code != 200 {
		t.Error("request after window reset should succeed")
	}
}

func TestMiddlewareErrorHandling(t *testing.T) {
	engine := New(nil)

	errorCaught := false
	engine.SetErrorHandler(func(req *Request, res *Response, err error) {
		errorCaught = true
		res.JSON(500, map[string]string{"error": err.Error()})
	})

	engine.Use(func(req *Request, res *Response) error {
		return req.Next()
	})

	engine.GET("/error", func(req *Request, res *Response) error {
		return fmt.Errorf("handler error")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/error", nil)

	engine.ServeHTTP(w, r)

	if !errorCaught {
		t.Error("expected error handler to be called")
	}
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestStatsMiddleware(t *testing.T) {
	stats := Stats()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	sc := GetStatsCollector()
	initialCount := sc.GetStats().RequestCount

	req.index = -1
	req.handlers = []HandlerFunc{
		func(req *Request, res *Response) error {
			return res.String(200, "ok")
		},
	}

	err := stats(req, res)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	finalCount := sc.GetStats().RequestCount
	if finalCount != initialCount+1 {
		t.Errorf("expected request count to increase from %d to %d, got %d", initialCount, initialCount+1, finalCount)
	}
}
