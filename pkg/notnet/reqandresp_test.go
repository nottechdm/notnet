package notnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRequestParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]string
		key      string
		expected string
	}{
		{"existing param", map[string]string{"id": "123"}, "id", "123"},
		{"non-existing param", map[string]string{"id": "123"}, "name", ""},
		{"multiple params", map[string]string{"id": "123", "name": "john"}, "name", "john"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			req.params = tt.params

			result := req.Param(tt.key)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRequestQuery(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		key      string
		expected string
	}{
		{"single query", "/test?q=hello", "q", "hello"},
		{"multiple queries", "/test?q=hello&page=1", "q", "hello"},
		{"multiple queries - second param", "/test?q=hello&page=1", "page", "1"},
		{"missing query", "/test?q=hello", "missing", ""},
		{"empty query", "/test", "q", ""},
		{"url encoded", "/test?q=hello%20world", "q", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.path, nil)
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			result := req.Query(tt.key)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRequestMethod(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, "/test", nil)
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			if req.Method() != method {
				t.Errorf("expected %s, got %s", method, req.Method())
			}
		})
	}
}

func TestRequestPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"root path", "/", "/"},
		{"simple path", "/users", "/users"},
		{"nested path", "/api/v1/users", "/api/v1/users"},
		{"path with query", "/users?id=1", "/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.path, nil)
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			if req.Path() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, req.Path())
			}
		})
	}
}

func TestRequestURL(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?q=hello", nil)
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	url := req.URL()
	if !strings.Contains(url, "/test") {
		t.Error("expected URL to contain path")
	}
	if !strings.Contains(url, "q=hello") {
		t.Error("expected URL to contain query")
	}
}

func TestRequestRemoteAddr(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "192.168.1.1:8000"
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	if req.RemoteAddr() != "192.168.1.1:8000" {
		t.Error("expected RemoteAddr to be returned")
	}
}

func TestRequestHeader(t *testing.T) {
	tests := []struct {
		name     string
		headKey  string
		headVal  string
		getKey   string
		expected string
	}{
		{"existing header", "Content-Type", "application/json", "Content-Type", "application/json"},
		{"missing header", "Content-Type", "application/json", "Authorization", ""},
		{"case insensitive", "Authorization", "Bearer token", "authorization", "Bearer token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			if tt.headVal != "" {
				r.Header.Set(tt.headKey, tt.headVal)
			}
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			result := req.Header(tt.getKey)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRequestBindJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
		check   func(map[string]interface{}) bool
	}{
		{
			name:    "valid json",
			body:    `{"name":"john","age":30}`,
			wantErr: false,
			check: func(data map[string]interface{}) bool {
				return data["name"] == "john" && data["age"] == float64(30)
			},
		},
		{
			name:    "invalid json",
			body:    `{invalid}`,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "empty json",
			body:    `{}`,
			wantErr: false,
			check: func(data map[string]interface{}) bool {
				return len(data) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req, res := AcquireRequestResponse(w, r)
			defer ReleaseRequestResponse(req, res)

			var data map[string]interface{}
			err := req.BindJSON(&data)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}

			if !tt.wantErr && tt.check != nil {
				if !tt.check(data) {
					t.Errorf("check failed for body: %s", tt.body)
				}
			}
		})
	}
}

func TestResponseString(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		content     string
		contentType string
	}{
		{"200 OK", 200, "ok", "text/plain; charset=utf-8"},
		{"201 Created", 201, "created", "text/plain; charset=utf-8"},
		{"404 Not Found", 404, "not found", "text/plain; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			_, res := AcquireRequestResponse(w, r)

			res.String(tt.status, tt.content)

			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, w.Code)
			}
			if w.Body.String() != tt.content {
				t.Errorf("expected body %s, got %s", tt.content, w.Body.String())
			}
			if w.Header().Get("Content-Type") != tt.contentType {
				t.Errorf("expected content type %s, got %s", tt.contentType, w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestResponseJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	data := map[string]interface{}{"message": "hello", "status": 200}
	err := res.JSON(200, data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Error("expected JSON content type")
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "hello" {
		t.Error("expected JSON data to be encoded")
	}
}

func TestResponseHTML(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	html := "<h1>Hello</h1>"
	err := res.HTML(200, html)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("expected HTML content type")
	}

	if w.Body.String() != html {
		t.Errorf("expected body %s, got %s", html, w.Body.String())
	}
}

func TestResponseBytes(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	data := []byte{0x89, 0x50, 0x4E, 0x47}
	err := res.Bytes(200, "image/png", data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "image/png" {
		t.Error("expected image/png content type")
	}

	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Error("expected binary data to be written")
	}
}

func TestResponseSetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	res.SetHeader("X-Custom", "value1")
	res.SetHeader("X-Another", "value2")

	if w.Header().Get("X-Custom") != "value1" {
		t.Error("expected custom header to be set")
	}
	if w.Header().Get("X-Another") != "value2" {
		t.Error("expected another header to be set")
	}
}

func TestResponseStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	res.Status(201)
	res.String(201, "created")

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestResponseRedirect(t *testing.T) {
	tests := []struct {
		code    int
		url     string
		wantErr bool
	}{
		{301, "/new", false},
		{302, "https://example.com", false},
		{307, "/temporary", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("redirect_%d", tt.code), func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/old", nil)

			_, res := AcquireRequestResponse(w, r)
			err := res.Redirect(tt.code, tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}
			if w.Code != tt.code {
				t.Errorf("expected status %d, got %d", tt.code, w.Code)
			}

			if loc := w.Header().Get("Location"); loc != tt.url {
				t.Errorf("expected Location %q, got %q", tt.url, loc)
			}
		})
	}
}

func TestRequestNext(t *testing.T) {
	calls := []int{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, _ := AcquireRequestResponse(w, r)

	req.handlers = []HandlerFunc{
		func(req *Request, res *Response) error {
			calls = append(calls, 1)
			return req.Next()
		},
		func(req *Request, res *Response) error {
			calls = append(calls, 2)
			return req.Next()
		},
		func(req *Request, res *Response) error {
			calls = append(calls, 3)
			return req.Next()
		},
	}
	req.index = -1

	req.Next()

	if len(calls) != 3 {
		t.Errorf("expected 3 calls, got %d", len(calls))
	}
	if calls[0] != 1 || calls[1] != 2 || calls[2] != 3 {
		t.Errorf("expected sequential calls, got %v", calls)
	}
}

func TestRequestNextEnd(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	req, _ := AcquireRequestResponse(w, r)

	req.handlers = []HandlerFunc{}
	req.index = -1

	err := req.Next()
	if err != nil {
		t.Errorf("expected no error at end, got %v", err)
	}
}

func TestPoolAcquireRelease(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	req, res := AcquireRequestResponse(w, r)

	if req == nil {
		t.Error("expected request to be acquired")
	}
	if res == nil {
		t.Error("expected response to be acquired")
	}

	ReleaseRequestResponse(req, res)

	// Acquire again to verify it was returned to pool
	req2, res2 := AcquireRequestResponse(w, r)

	if req2 == nil {
		t.Error("expected request to be reused from pool")
	}
	if res2 == nil {
		t.Error("expected response to be reused from pool")
	}
}

func TestResponseSetHeaderChaining(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	result := res.SetHeader("X-A", "a").SetHeader("X-B", "b").SetHeader("X-C", "c")

	if result != res {
		t.Error("expected SetHeader to return response for chaining")
	}

	if w.Header().Get("X-A") != "a" {
		t.Error("expected X-A header")
	}
	if w.Header().Get("X-B") != "b" {
		t.Error("expected X-B header")
	}
	if w.Header().Get("X-C") != "c" {
		t.Error("expected X-C header")
	}
}

func TestResponseStatusChaining(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	result := res.Status(201)

	if result != res {
		t.Error("expected Status to return response for chaining")
	}
}

func TestComplexJSONStructure(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(`{
		"user": {
			"name": "john",
			"age": 30,
			"tags": ["admin", "user"]
		},
		"active": true
	}`))
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	type User struct {
		Name string   `json:"name"`
		Age  int      `json:"age"`
		Tags []string `json:"tags"`
	}
	type Data struct {
		User   User `json:"user"`
		Active bool `json:"active"`
	}

	var data Data
	err := req.BindJSON(&data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if data.User.Name != "john" {
		t.Error("expected name to be john")
	}
	if data.User.Age != 30 {
		t.Error("expected age to be 30")
	}
	if len(data.User.Tags) != 2 {
		t.Error("expected 2 tags")
	}
	if !data.Active {
		t.Error("expected active to be true")
	}
}

func TestMultipleHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Add("Cookie", "session=abc")
	r.Header.Add("Cookie", "lang=en")
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	// Get should return only the first value
	cookie := req.Header("Cookie")
	if cookie == "" {
		t.Error("expected cookie header")
	}
}

func TestEmptyResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	err := res.String(204, "")

	if err != nil {
		t.Errorf("expected no error for empty response, got %v", err)
	}

	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Body.String() != "" {
		t.Error("expected empty body")
	}
}

func TestResponseFile(t *testing.T) {
	// Create a temporary file
	content := "hello from file"
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	_, res := AcquireRequestResponse(w, r)

	err = res.File(tmpfile.Name())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if w.Body.String() != content {
		t.Errorf("expected body %s, got %s", content, w.Body.String())
	}
}
