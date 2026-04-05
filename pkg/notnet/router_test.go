package notnet

import (
	"testing"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("expected router to be created")
	}
	if router.tree == nil {
		t.Fatal("expected tree to be created")
	}
	if len(router.tree) != 0 {
		t.Error("expected tree to be empty initially")
	}
}

func TestRouterRegister(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name    string
		method  string
		path    string
		handler HandlerFunc
	}{
		{"GET /users", "GET", "/users", func(req *Request, res *Response) error { return nil }},
		{"POST /users", "POST", "/users", func(req *Request, res *Response) error { return nil }},
		{"GET /users/:id", "GET", "/users/:id", func(req *Request, res *Response) error { return nil }},
		{"PUT /users/:id", "PUT", "/users/:id", func(req *Request, res *Response) error { return nil }},
		{"DELETE /users/:id", "DELETE", "/users/:id", func(req *Request, res *Response) error { return nil }},
		{"GET /posts/:id/comments/:comment_id", "GET", "/posts/:id/comments/:comment_id", func(req *Request, res *Response) error { return nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router.Register(tt.method, tt.path, tt.handler)
			key := tt.method + ":" + tt.path
			if node, ok := router.tree[key]; !ok || node.handler == nil {
				t.Error("expected route to be registered")
			}
		})
	}
}

func TestRouterRegisterOverwrite(t *testing.T) {
	router := NewRouter()
	handler1 := func(req *Request, res *Response) error { return nil }
	handler2 := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/test", handler1)
	router.Register("GET", "/test", handler2)

	route, _, found := router.Match("GET", "/test")
	if !found {
		t.Error("expected route to be found after overwrite")
	}
	if route == nil {
		t.Error("expected route to be not nil")
	}
}

func TestRouterMatchExact(t *testing.T) {
	router := NewRouter()
	handler := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/users", handler)

	tests := []struct {
		name    string
		method  string
		path    string
		found   bool
		wantURL string
	}{
		{"exact match", "GET", "/users", true, "/users"},
		{"no match - different path", "GET", "/posts", false, ""},
		{"no match - different method", "POST", "/users", false, ""},
		{"no match - case sensitive", "GET", "/Users", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, _, found := router.Match(tt.method, tt.path)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if found && route.Path != tt.wantURL {
				t.Errorf("expected URL %s, got %s", tt.wantURL, route.Path)
			}
		})
	}
}

func TestRouterMatchWithParams(t *testing.T) {
	router := NewRouter()
	handler := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/users/:id", handler)

	tests := []struct {
		name      string
		method    string
		path      string
		found     bool
		paramName string
		paramVal  string
	}{
		{"param match", "GET", "/users/123", true, "id", "123"},
		{"param match with string", "GET", "/users/john", true, "id", "john"},
		{"no match - more segments", "GET", "/users/123/posts", false, "", ""},
		{"no match - less segments", "GET", "/users", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, params, found := router.Match(tt.method, tt.path)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if found && params[tt.paramName] != tt.paramVal {
				t.Errorf("expected param %s=%s, got %s", tt.paramName, tt.paramVal, params[tt.paramName])
			}
		})
	}
}

func TestRouterMatchMultipleParams(t *testing.T) {
	router := NewRouter()
	handler := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/posts/:id/comments/:comment_id", handler)

	route, params, found := router.Match("GET", "/posts/42/comments/99")

	if !found {
		t.Error("expected route to be found")
	}
	if params["id"] != "42" {
		t.Errorf("expected id=42, got %s", params["id"])
	}
	if params["comment_id"] != "99" {
		t.Errorf("expected comment_id=99, got %s", params["comment_id"])
	}
	if route.Handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestRouterMatchPriority(t *testing.T) {
	router := NewRouter()
	exactHandler := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/users/me", exactHandler)

	_, params, found := router.Match("GET", "/users/me")

	if !found {
		t.Error("expected exact match to be found first")
	}
	if len(params) != 0 {
		t.Errorf("expected no params for exact match, got %d", len(params))
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		matches bool
		params  map[string]string
	}{
		{
			"exact match",
			"/users",
			"/users",
			true,
			map[string]string{},
		},
		{
			"single param",
			"/users/:id",
			"/users/123",
			true,
			map[string]string{"id": "123"},
		},
		{
			"multiple params",
			"/posts/:id/comments/:comment_id",
			"/posts/1/comments/2",
			true,
			map[string]string{"id": "1", "comment_id": "2"},
		},
		{
			"no match - different segments",
			"/users/:id",
			"/users/123/posts",
			false,
			map[string]string{},
		},
		{
			"no match - fewer segments",
			"/users/:id",
			"/users",
			false,
			map[string]string{},
		},
		{
			"no match - partial segment",
			"/users/:id",
			"/user/123",
			false,
			map[string]string{},
		},
		{
			"param with alphanumeric",
			"/api/:version",
			"/api/v1",
			true,
			map[string]string{"version": "v1"},
		},
		{
			"param with special chars",
			"/files/:name",
			"/files/my-file.txt",
			true,
			map[string]string{"name": "my-file.txt"},
		},
		{
			"trailing slash mismatch",
			"/users/",
			"/users",
			false,
			map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, params := matchPath(tt.pattern, tt.path)

			if matches != tt.matches {
				t.Errorf("expected matches=%v, got %v", tt.matches, matches)
			}

			if matches {
				for k, v := range tt.params {
					if params[k] != v {
						t.Errorf("expected param %s=%s, got %s", k, v, params[k])
					}
				}
			}
		})
	}
}

func TestRadixNode(t *testing.T) {
	node := &RadixNode{
		edge:     "/test",
		param:    "id",
		isParam:  true,
		children: make(map[string]*RadixNode),
	}

	if node.edge != "/test" {
		t.Error("expected edge to be /test")
	}
	if node.param != "id" {
		t.Error("expected param to be id")
	}
	if !node.isParam {
		t.Error("expected isParam to be true")
	}
	if node.children == nil {
		t.Error("expected children map to be created")
	}
}

func TestRouterComplex(t *testing.T) {
	router := NewRouter()

	routes := []struct {
		method string
		path   string
		name   string
	}{
		{"GET", "/", "index"},
		{"GET", "/users", "list users"},
		{"GET", "/users/:id", "get user"},
		{"POST", "/users", "create user"},
		{"PUT", "/users/:id", "update user"},
		{"DELETE", "/users/:id", "delete user"},
		{"GET", "/users/:id/posts", "list user posts"},
		{"GET", "/users/:id/posts/:post_id", "get user post"},
		{"POST", "/users/:id/posts", "create user post"},
		{"PUT", "/users/:id/posts/:post_id", "update user post"},
		{"DELETE", "/users/:id/posts/:post_id", "delete user post"},
	}

	handler := func(req *Request, res *Response) error { return nil }

	for _, r := range routes {
		router.Register(r.method, r.path, handler)
	}

	// Test all routes
	for _, r := range routes {
		t.Run(r.name, func(t *testing.T) {
			switch r.path {
			case "/":
				_, _, found := router.Match(r.method, "/")
				if !found {
					t.Error("expected route to be found")
				}
			case "/users":
				_, _, found := router.Match(r.method, "/users")
				if !found {
					t.Error("expected route to be found")
				}
			case "/users/:id":
				_, params, found := router.Match(r.method, "/users/42")
				if !found {
					t.Error("expected route to be found")
				}
				if params["id"] != "42" {
					t.Errorf("expected id=42, got %s", params["id"])
				}
			case "/users/:id/posts":
				_, params, found := router.Match(r.method, "/users/42/posts")
				if !found {
					t.Error("expected route to be found")
				}
				if params["id"] != "42" {
					t.Errorf("expected id=42, got %s", params["id"])
				}
			case "/users/:id/posts/:post_id":
				_, params, found := router.Match(r.method, "/users/42/posts/99")
				if !found {
					t.Error("expected route to be found")
				}
				if params["id"] != "42" {
					t.Errorf("expected id=42, got %s", params["id"])
				}
				if params["post_id"] != "99" {
					t.Errorf("expected post_id=99, got %s", params["post_id"])
				}
			}
		})
	}
}

func TestRouterEmptyPath(t *testing.T) {
	router := NewRouter()
	handler := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/", handler)
	route, _, found := router.Match("GET", "/")

	if !found {
		t.Error("expected root path to be found")
	}
	if route == nil {
		t.Error("expected route to be not nil")
	}
}

func TestRouterLongPath(t *testing.T) {
	router := NewRouter()
	handler := func(req *Request, res *Response) error { return nil }

	longPath := "/api/v1/users/:user_id/posts/:post_id/comments/:comment_id/replies/:reply_id"
	router.Register("GET", longPath, handler)

	testPath := "/api/v1/users/10/posts/20/comments/30/replies/40"
	route, params, found := router.Match("GET", testPath)

	if !found {
		t.Error("expected long path to be found")
	}
	if route == nil {
		t.Error("expected route to be not nil")
	}
	if params["user_id"] != "10" {
		t.Error("expected user_id=10")
	}
	if params["post_id"] != "20" {
		t.Error("expected post_id=20")
	}
	if params["comment_id"] != "30" {
		t.Error("expected comment_id=30")
	}
	if params["reply_id"] != "40" {
		t.Error("expected reply_id=40")
	}
}

func TestRouterDifferentMethods(t *testing.T) {
	router := NewRouter()
	handler := func(req *Request, res *Response) error { return nil }

	router.Register("GET", "/items", handler)
	router.Register("POST", "/items", handler)
	router.Register("PUT", "/items/:id", handler)
	router.Register("DELETE", "/items/:id", handler)

	tests := []struct {
		method string
		path   string
		found  bool
	}{
		{"GET", "/items", true},
		{"POST", "/items", true},
		{"PUT", "/items/1", true},
		{"DELETE", "/items/1", true},
		{"PATCH", "/items/1", false},
		{"POST", "/items/1", false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			_, _, found := router.Match(tt.method, tt.path)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
		})
	}
}
