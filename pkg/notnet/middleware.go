package notnet

import (
	"fmt"
	"strings"
	"time"
)

// Recovery middleware catches panics
// It uses a defer function to recover from panics and returns a 500 Internal Server Error response
// Example usage: app.Use(notnet.Recovery())
func Recovery() HandlerFunc {
	return func(req *Request, res *Response) error {
		defer func() {
			// Recover from panic and log the error
			if rec := recover(); rec != nil {
				fmt.Printf("panic: %v\n", rec)
				// Return a generic error response to the client
				res.JSON(500, map[string]string{"error": "internal server error"})
			}
		}()
		// Call the next handler in the chain
		return req.Next()
	}
}

// Logger middleware logs requests
// It logs the method, path, remote address, and processing time
// Log format: [METHOD] PATH REMOTE_ADDR - ELAPSED_MSms
// Example: [GET] /hello
func Logger() HandlerFunc {
	return func(req *Request, res *Response) error {
		start := time.Now()
		err := req.Next()
		elapsed := time.Since(start)
		// Format log message with method, path, remote address, and elapsed time
		fmt.Printf("[%s] %s %s - %dms\n", req.Method(), req.Path(), req.HTTPRequest.RemoteAddr, elapsed.Milliseconds())
		return err
	}
}

// AuthRequired middleware requires authentication
// It checks for the presence of an Authorization header
// If the header is missing, it returns a 401 Unauthorized response
func AuthRequired() HandlerFunc {
	return func(req *Request, res *Response) error {
		token := req.Header("Authorization")
		if token == "" {
			// Return a 401 Unauthorized response if the Authorization header is missing
			res.JSON(401, map[string]string{"error": "unauthorized"})
			return fmt.Errorf("unauthorized")
		}
		return req.Next()
	}
}

// CORSConfig defines the configuration for CORS middleware
// AllowOrigins: list of allowed origins (e.g. ["*"] or ["https://example.com"])
// AllowMethods: list of allowed HTTP methods (e.g. ["GET", "POST"])
// AllowHeaders: list of allowed headers (e.g. ["Content-Type", "Authorization"])
// CustomHeaders: map of custom headers to add to the response
type CORSConfig struct {
	// Defaults to allowing all origins, methods, and headers if not specified
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
	// CustomHeaders allows users to add any additional headers to the response
	CustomHeaders map[string]string
}

// CORS middleware enables CORS
func CORS(config *CORSConfig) HandlerFunc {
	if config == nil {
		// Set default CORS configuration if none provided
		config = &CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders: []string{"Content-Type", "Authorization"},
		}
	}

	// Set defaults for any missing fields
	if config.AllowHeaders == nil {
		config.AllowHeaders = []string{"Content-Type", "Authorization"}
	}
	if config.AllowMethods == nil {
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}
	if config.AllowOrigins == nil {
		config.AllowOrigins = []string{"*"}
	}

	return func(req *Request, res *Response) error {
		res.SetHeader("Access-Control-Allow-Origin", strings.Join(config.AllowOrigins, ", "))
		res.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		res.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		// Custom Headers
		if config.CustomHeaders != nil {
			for k, v := range config.CustomHeaders {
				// Set any additional custom headers specified in the configuration
				res.SetHeader(k, v)
			}
		}
		// Handle preflight OPTIONS request
		if req.Method() == "OPTIONS" {
			// For preflight requests, return a 204 No Content response with the appropriate CORS headers
			return res.Status(204).String(204, "")
		}

		return req.Next()
	}
}

// RateLimit creates a simple rate limit middleware
// maxRequests: maximum number of requests allowed within the window duration
// window: time duration for the rate limit (e.g. 1 minute)
// It uses an in-memory map to track request counts per remote address
func RateLimit(maxRequests int, window time.Duration) HandlerFunc {
	type limiter struct {
		// Count of requests made by the remote address within the current window
		count int
		// Reset time for the current window
		reset time.Time
	}
	// limiters map tracks the limiter for each remote address
	limiters := make(map[string]*limiter)

	return func(req *Request, res *Response) error {
		addr := req.RemoteAddr()
		now := time.Now()

		l, exists := limiters[addr]
		if !exists || now.After(l.reset) {
			limiters[addr] = &limiter{count: 1, reset: now.Add(window)}
		} else {
			l.count++
			if l.count > maxRequests {
				return res.JSON(429, map[string]string{"error": "rate limit exceeded"})
			}
		}

		return req.Next()
	}
}

// RequestID middleware adds a request ID
// It checks for a header (e.g. "X-Request-ID") and uses it as the request ID
// If the header is missing, it generates a new request ID using the current timestamp
// It sets the request ID in the response header (e.g. "X-Request-ID")
func RequestID(headerName string) HandlerFunc {
	return func(req *Request, res *Response) error {
		id := req.Header(headerName)
		if id == "" {
			id = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		res.SetHeader(headerName, id)
		return req.Next()
	}
}

// Stats middleware collects request statistics for the dashboard
func Stats() HandlerFunc {
	return func(req *Request, res *Response) error {
		start := time.Now()
		err := req.Next()
		elapsed := time.Since(start)
		// Record the request in the stats collector
		GetStatsCollector().RecordRequest(elapsed)
		return err
	}
}
