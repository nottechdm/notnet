package notnet

import (
	"log"
	"net"
	"net/http"
	"time"
)

// HandlerFunc is the request handler function type
type HandlerFunc func(*Request, *Response) error

// Engine is the core NotNet application
//
// It implements http.Handler and manages routing, middleware, and error handling
// It provides methods for defining routes, groups, and starting the server
// It also allows users to set custom error handlers and panic handlers
// The Engine struct contains the router, middleware stack, and server configuration
type Engine struct {
	// Core components
	router *Router
	// Middleware and handlers
	middleware []HandlerFunc
	srv        *http.Server
	// Error handling
	errorFunc    ErrorHandlerFunc
	notFoundFunc NotFoundHandlerFunc
	panicFunc    PanicHandlerFunc

	// Tuning
	maxHeaderBytes int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
}

// ErrorHandlerFunc is called when a handler returns an error
type ErrorHandlerFunc func(*Request, *Response, error)

// NotFoundHandlerFunc is called when no route matches
type NotFoundHandlerFunc func(*Request, *Response)

// PanicHandlerFunc is called when a panic occurs
type PanicHandlerFunc func(*Request, *Response, interface{})

// EngineOption defines options for creating a new Engine
//
// MaxHeaderBytes: maximum size of request headers (default 1MB)
// ReadTimeout: maximum duration for reading the entire request (default 15s)
// WriteTimeout: maximum duration before timing out writes of the response (default 15s)
// IdleTimeout: maximum amount of time to wait for the next request when keep-alives are enabled (default 60s)
// ErrorFunc: custom error handler function
// NotFoundFunc: custom 404 handler function
// PanicFunc: custom panic handler function
type EngineOption struct {
	// Header options
	MaxHeaderBytes int
	// Timeout settings for the server
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// Error handling functions
	ErrorFunc    *ErrorHandlerFunc
	NotFoundFunc *NotFoundHandlerFunc
	PanicFunc    *PanicHandlerFunc
}

// New creates a new NotNet engine
//
// It takes an optional EngineOption struct to configure the engine's settings and handlers
// If no options are provided, it uses default settings for timeouts and error handlers
// Example usage:
//   errHandler := notnet.ErrorHandlerFunc(customErrorHandler)
//   app := notnet.New(&notnet.EngineOption{
//       ReadTimeout: 10 * time.Second,
//       ErrorFunc:   &errHandler,
//   })
func New(opts *EngineOption) *Engine {
	e := &Engine{
		router:         NewRouter(),
		middleware:     make([]HandlerFunc, 0),
		maxHeaderBytes: 1 << 20, // 1MB
		readTimeout:    15 * time.Second,
		writeTimeout:   15 * time.Second,
		idleTimeout:    60 * time.Second,
	}

	// Override defaults with options
	if opts != nil {
		if opts.MaxHeaderBytes > 0 {
			e.maxHeaderBytes = opts.MaxHeaderBytes
		}
		if opts.ReadTimeout > 0 {
			e.readTimeout = opts.ReadTimeout
		}
		if opts.WriteTimeout > 0 {
			e.writeTimeout = opts.WriteTimeout
		}
		if opts.IdleTimeout > 0 {
			e.idleTimeout = opts.IdleTimeout
		}
		if opts.ErrorFunc != nil {
			e.errorFunc = *opts.ErrorFunc
		}
		if opts.NotFoundFunc != nil {
			e.notFoundFunc = *opts.NotFoundFunc
		}
		if opts.PanicFunc != nil {
			e.panicFunc = *opts.PanicFunc
		}
	}

	// Set default handlers if not provided
	e.errorFunc = e.defaultErrorHandler
	e.notFoundFunc = e.defaultNotFoundHandler
	e.panicFunc = e.defaultPanicHandler

	return e
}

// Use adds global middleware to the engine
//
// Example usage: app.Use(notnet.Logger(), notnet.Recovery())
func (e *Engine) Use(handlers ...HandlerFunc) *Engine {
	e.middleware = append(e.middleware, handlers...)
	return e
}

// GET registers a GET route
//
// It takes a path and a handler function, and registers it with the router for the GET method
// Example usage: app.GET("/ping", func(req *notnet.Request, res *notnet.Response) error { return res.String(200, "pong") })
func (e *Engine) GET(path string, handler HandlerFunc) *Engine {
	e.router.Register("GET", path, handler)
	return e
}

// POST registers a POST route
//
// It takes a path and a handler function, and registers it with the router for the POST method
// Example usage: app.POST("/users", func(req *notnet.Request, res *notnet.Response) error { return res.String(201, "user created") })
func (e *Engine) POST(path string, handler HandlerFunc) *Engine {
	e.router.Register("POST", path, handler)
	return e
}

// PUT registers a PUT route
//
// It takes a path and a handler function, and registers it with the router for the PUT method
// Example usage: app.PUT("/users/1", func(req *notnet.Request, res *notnet.Response) error { return res.String(200, "user updated") })
func (e *Engine) PUT(path string, handler HandlerFunc) *Engine {
	e.router.Register("PUT", path, handler)
	return e
}

// DELETE registers a DELETE route
//
// It takes a path and a handler function, and registers it with the router for the DELETE method
// Example usage: app.DELETE("/users/1", func(req *notnet.Request, res *notnet.Response) error { return res.String(204, "") })
func (e *Engine) DELETE(path string, handler HandlerFunc) *Engine {
	e.router.Register("DELETE", path, handler)
	return e
}

// PATCH registers a PATCH route
//
// It takes a path and a handler function, and registers it with the router for the PATCH method
// Example usage: app.PATCH("/users/1", func(req *notnet.Request, res *notnet.Response) error { return res.String(200, "user partially updated") })
func (e *Engine) PATCH(path string, handler HandlerFunc) *Engine {
	e.router.Register("PATCH", path, handler)
	return e
}

// OPTIONS registers an OPTIONS route
//
// It takes a path and a handler function, and registers it with the router for the OPTIONS method
// Example usage: app.OPTIONS("/users", func(req *notnet.Request, res *notnet.Response) error { return res.String(204, "") })
func (e *Engine) OPTIONS(path string, handler HandlerFunc) *Engine {
	e.router.Register("OPTIONS", path, handler)
	return e
}

// HEAD registers a HEAD route
//
// It takes a path and a handler function, and registers it with the router for the HEAD method
// Example usage: app.HEAD("/users/1", func(req *notnet.Request, res *notnet.Response) error { return res.String(200, "") })
func (e *Engine) HEAD(path string, handler HandlerFunc) *Engine {
	e.router.Register("HEAD", path, handler)
	return e
}

// ApplyConfig applies configuration options to the engine.
//
// It takes an EngineOption struct and updates engine-wide configuration accordingly.
// Because route registration methods such as GET/POST/PUT return *Engine, chaining
// ApplyConfig after registering a route still updates the shared engine, not only
// the route that was just added.
// Example usage: app.ApplyConfig(&notnet.EngineOption{ReadTimeout: 10 * time.Second})
func (e *Engine) ApplyConfig(config *EngineOption) *Engine {
	// Override defaults with options
	if config != nil {
		if config.MaxHeaderBytes > 0 {
			e.maxHeaderBytes = config.MaxHeaderBytes
		}
		if config.ReadTimeout > 0 {
			e.readTimeout = config.ReadTimeout
		}
		if config.WriteTimeout > 0 {
			e.writeTimeout = config.WriteTimeout
		}
		if config.IdleTimeout > 0 {
			e.idleTimeout = config.IdleTimeout
		}
		if config.ErrorFunc != nil {
			e.errorFunc = *config.ErrorFunc
		}
		if config.NotFoundFunc != nil {
			e.notFoundFunc = *config.NotFoundFunc
		}
		if config.PanicFunc != nil {
			e.panicFunc = *config.PanicFunc
		}
	}
	return e
}

// Group creates a route group with optional middleware
//
// It takes a path prefix and optional middleware, and returns a Group object for defining routes within that group
// Example usage: api := app.Group("/api/v1", notnet.AuthRequired()); api.GET("/status", func(req *notnet.Request, res *notnet.Response) error { return res.JSON(200, map[string]string{"status": "ok"}) })
func (e *Engine) Group(path string, handlers ...HandlerFunc) *Group {
	return &Group{
		prefix:   path,
		engine:   e,
		handlers: handlers,
	}
}

// SetErrorHandler sets custom error handler
//
// It takes a function that will be called whenever a handler returns an error
// Example usage: app.SetErrorHandler(func(req *notnet.Request, res *notnet.Response, err error) { log.Printf("error: %v", err); res.JSON(500, map[string]string{"error": err.Error()}) })
func (e *Engine) SetErrorHandler(f ErrorHandlerFunc) *Engine {
	e.errorFunc = f
	return e
}

// SetNotFoundHandler sets custom 404 handler
//
// It takes a function that will be called whenever no route matches the incoming request
// Example usage: app.SetNotFoundHandler(func(req *notnet.Request, res *notnet.Response) { res.JSON(404, map[string]string{"error": "endpoint not found", "path": req.Path()}) })
func (e *Engine) SetNotFoundHandler(f NotFoundHandlerFunc) *Engine {
	e.notFoundFunc = f
	return e
}

// SetPanicHandler sets custom panic handler
// It takes a function that will be called whenever a panic occurs during request handling
// Example usage: app.SetPanicHandler(func(req *notnet.Request, res *notnet.Response, rec interface{}) { log.Printf("panic: %v", rec); res.JSON(500, map[string]string{"error": "internal server error"}) })
func (e *Engine) SetPanicHandler(f PanicHandlerFunc) *Engine {
	e.panicFunc = f
	return e
}

// Listen starts the server on the given address
//
// It creates an http.Server with the configured timeouts and handlers, and starts listening for incoming requests
// Example usage: app.Listen(":8080")
// Note: For production use, consider using ListenTLS with proper TLS certificates
// Note: Listen will block the current goroutine, so it should be called in the main function or a separate goroutine
func (e *Engine) Listen(addr string) error {
	e.srv = &http.Server{
		// Set the address and handler for the server
		// Addr specifies the TCP address to listen on, in the form "host:port"
		Addr: addr,
		// The Engine itself implements http.Handler, so we set it as the handler for incoming requests
		// This allows the Engine to handle routing, middleware, and error handling for all requests
		Handler: e,
		// Configure server timeouts and limits based on the Engine's settings
		MaxHeaderBytes: e.maxHeaderBytes,
		// ReadTimeout is the maximum duration for reading the entire request, including the body
		ReadTimeout: e.readTimeout,
		// WriteTimeout is the maximum duration before timing out writes of the response
		WriteTimeout: e.writeTimeout,
		// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled
		IdleTimeout: e.idleTimeout,
	}
	return e.srv.ListenAndServe()
}

// ListenTLS starts a TLS server
//
// It creates an http.Server with the configured timeouts and handlers, and starts listening for incoming requests over TLS
// Example usage: app.ListenTLS(":8443", "cert.pem", "key.pem")
// Note: certFile and keyFile should be valid TLS certificate and key files for the server to start successfully
func (e *Engine) ListenTLS(addr string, certFile, keyFile string) error {
	e.srv = &http.Server{
		// Set the address and handler for the server
		// Addr specifies the TCP address to listen on, in the form "host:port"
		Addr: addr,
		// The Engine itself implements http.Handler, so we set it as the handler for incoming requests
		// This allows the Engine to handle routing, middleware, and error handling for all requests
		Handler: e,
		// Configure server timeouts and limits based on the Engine's settings
		MaxHeaderBytes: e.maxHeaderBytes,
		ReadTimeout:    e.readTimeout,
		WriteTimeout:   e.writeTimeout,
		IdleTimeout:    e.idleTimeout,
	}
	return e.srv.ListenAndServeTLS(certFile, keyFile)
}

// ListenListener uses a custom listener
//
// It allows users to create their own net.Listener (e.g. for Unix sockets or custom TCP listeners) and start the server with it
// Example usage: listener, _ := net.Listen("tcp", ":8080"); app.ListenListener(listener)
// Note: The provided listener should be properly configured and ready to accept connections for the server to start successfully
func (e *Engine) ListenListener(listener net.Listener) error {
	e.srv = &http.Server{
		// Set the handler for the server
		// The Engine itself implements http.Handler, so we set it as the handler for incoming requests
		// This allows the Engine to handle routing, middleware, and error handling for all requests
		Handler: e,
		// Configure server timeouts and limits based on the Engine's settings
		MaxHeaderBytes: e.maxHeaderBytes,
		ReadTimeout:    e.readTimeout,
		WriteTimeout:   e.writeTimeout,
		IdleTimeout:    e.idleTimeout,
	}
	return e.srv.Serve(listener)
}

// Shutdown gracefully shuts down the server
//
// It closes the server and releases any resources associated with it
// Example usage: app.Shutdown()
// Note: Shutdown will close the server and stop accepting new connections, but it does not wait for existing connections to finish. For a more graceful shutdown, consider implementing a context with timeout and using srv.Shutdown(ctx) instead.
func (e *Engine) Shutdown() error {
	if e.srv != nil {
		// Close the server and release any resources associated with it
		return e.srv.Close()
	}
	return nil
}

// ServeHTTP implements http.Handler
// It is called for each incoming HTTP request and is responsible for routing the request, executing middleware, and handling errors
// The method first acquires a Request and Response object from the pool, and defers their release back to the pool
// It then uses the router to match the incoming request to a registered route, and if found, builds the handler chain with global middleware and the route handler
// If no route matches, it calls the notFoundFunc to return a 404 response
// If a panic occurs during request handling, it recovers and calls the panicFunc to return a 500 response
// Finally, it executes the handler chain by calling req.Next(), and if any handler returns an error, it calls the errorFunc to handle it
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, res := AcquireRequestResponse(w, r)
	defer ReleaseRequestResponse(req, res)

	defer func() {
		if rec := recover(); rec != nil {
			e.panicFunc(req, res, rec)
		}
	}()

	// Match route
	route, params, found := e.router.Match(r.Method, r.URL.Path)
	if !found {
		e.notFoundFunc(req, res)
		return
	}

	req.params = params

	// Build handler chain
	handlers := make([]HandlerFunc, 0, len(e.middleware)+1)
	handlers = append(handlers, e.middleware...)
	handlers = append(handlers, route.Handler)

	req.handlers = handlers
	req.index = -1

	// Execute chain
	if err := req.Next(); err != nil {
		e.errorFunc(req, res, err)
	}
}

// defaultErrorHandler handles errors
func (e *Engine) defaultErrorHandler(req *Request, res *Response, err error) {
	log.Printf("error: %v", err)
	if res.Writer.Header().Get("Content-Type") != "" {
		return
	}
	res.JSON(500, map[string]interface{}{
		"error": err.Error(),
	})
}

// defaultNotFoundHandler handles 404
func (e *Engine) defaultNotFoundHandler(req *Request, res *Response) {
	res.JSON(404, map[string]string{
		"error": "route not found",
	})
}

// defaultPanicHandler handles panics
func (e *Engine) defaultPanicHandler(req *Request, res *Response, rec interface{}) {
	log.Printf("panic: %v", rec)
	res.JSON(500, map[string]interface{}{
		"error": "internal server error",
	})
}

// Group represents a group of routes
// It allows users to define a common path prefix and shared middleware for a set of routes
// The Group struct contains the prefix, reference to the Engine, and any middleware handlers that should be applied to all routes in the group
// It provides methods for defining routes within the group (GET, POST, etc.) that automatically include the group's prefix and middleware
type Group struct {
	// prefix is the common path prefix for all routes in the group (e.g. "/api/v1")
	prefix string
	// engine is a reference to the main Engine, used to register routes and access shared functionality
	engine *Engine
	// handlers is the list of middleware handlers that should be applied to all routes in the group (e.g. authentication middleware)
	handlers []HandlerFunc
}

// Use adds middleware to the group
// It takes one or more HandlerFunc and appends them to the group's middleware stack
// Example usage: api := app.Group("/api/v1", notnet.AuthRequired()); api.Use(notnet.Logger())
func (g *Group) Use(handlers ...HandlerFunc) *Group {
	g.handlers = append(g.handlers, handlers...)
	return g
}

// GET registers a GET route in the group
func (g *Group) GET(path string, handler HandlerFunc) *Group {
	fullPath := g.prefix + path
	wrappedHandler := g.wrapHandler(handler)
	g.engine.router.Register("GET", fullPath, wrappedHandler)
	return g
}

// POST registers a POST route in the group
func (g *Group) POST(path string, handler HandlerFunc) *Group {
	fullPath := g.prefix + path
	wrappedHandler := g.wrapHandler(handler)
	g.engine.router.Register("POST", fullPath, wrappedHandler)
	return g
}

// PUT registers a PUT route in the group
func (g *Group) PUT(path string, handler HandlerFunc) *Group {
	fullPath := g.prefix + path
	wrappedHandler := g.wrapHandler(handler)
	g.engine.router.Register("PUT", fullPath, wrappedHandler)
	return g
}

// DELETE registers a DELETE route in the group
func (g *Group) DELETE(path string, handler HandlerFunc) *Group {
	fullPath := g.prefix + path
	wrappedHandler := g.wrapHandler(handler)
	g.engine.router.Register("DELETE", fullPath, wrappedHandler)
	return g
}

// PATCH registers a PATCH route in the group
func (g *Group) PATCH(path string, handler HandlerFunc) *Group {
	fullPath := g.prefix + path
	wrappedHandler := g.wrapHandler(handler)
	g.engine.router.Register("PATCH", fullPath, wrappedHandler)
	return g
}

// wrapHandler wraps the handler with group middleware
func (g *Group) wrapHandler(handler HandlerFunc) HandlerFunc {
	return func(req *Request, res *Response) error {
		// Save current handlers
		originalHandlers := req.handlers
		originalIndex := req.index

		// Create new handler chain with group middleware
		groupHandlers := make([]HandlerFunc, 0, len(g.handlers)+1)
		groupHandlers = append(groupHandlers, g.handlers...)
		groupHandlers = append(groupHandlers, handler)

		// Execute group middleware + handler
		req.handlers = groupHandlers
		req.index = -1
		err := req.Next()

		// Restore original state
		req.handlers = originalHandlers
		req.index = originalIndex

		return err
	}
}

// ApplyConfig applies configuration options to the underlying engine globally.
// It does not apply configuration only to this group or to the most recently
// registered route.
// Example usage: api.ApplyConfig(&notnet.EngineOption{ReadTimeout: 10 * time.Second})
func (g *Group) ApplyConfig(config *EngineOption) *Group {
	g.engine.ApplyConfig(config)
	return g
}
