package notnet

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// Response represents an HTTP response wrapper
// It provides methods for sending different types of responses (JSON, HTML, files, etc.) and setting headers
// The Response struct contains a reference to the http.ResponseWriter and the associated Request, allowing it to access request data and write responses
type Response struct {
	// Writer is the http.ResponseWriter used to send the response back to the client
	Writer http.ResponseWriter
	// req is a reference to the associated Request, allowing access to request data and context
	req *Request
}

// Request represents an HTTP request wrapper
// It provides methods for accessing request data (path parameters, query parameters, headers, etc.) and controlling the flow of middleware and handlers
// The Request struct contains a reference to the original http.Request, a reference to the associated Response, and fields for managing path parameters, middleware handlers, and the current handler index
type Request struct {
	// HTTPRequest is the original http.Request received from the client, containing all request data and context
	HTTPRequest *http.Request
	res         *Response
	params      map[string]string
	handlers    []HandlerFunc
	index       int
}

// Param returns a path parameter value
func (r *Request) Param(key string) string {
	return r.params[key]
}

// Query returns a query parameter value
func (r *Request) Query(key string) string {
	return r.HTTPRequest.URL.Query().Get(key)
}

// Next moves to the next handler in the chain
func (r *Request) Next() error {
	r.index++
	if r.index < len(r.handlers) {
		return r.handlers[r.index](r, r.res)
	}
	return nil
}

// Status sets the HTTP status code
func (r *Response) Status(code int) *Response {
	r.Writer.WriteHeader(code)
	return r
}

// String writes a string response
func (r *Response) String(code int, s string) error {
	r.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	r.Writer.WriteHeader(code)
	if len(s) > 0 {
		_, err := io.WriteString(r.Writer, s)
		return err
	}
	return nil
}

// JSON writes a JSON response
func (r *Response) JSON(code int, data interface{}) error {
	r.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Writer.WriteHeader(code)
	return json.NewEncoder(r.Writer).Encode(data)
}

// BindJSON parses the request body as JSON
func (r *Request) BindJSON(v interface{}) error {
	return json.NewDecoder(r.HTTPRequest.Body).Decode(v)
}

// HTML writes an HTML response
func (r *Response) HTML(code int, html string) error {
	r.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Writer.WriteHeader(code)
	_, err := io.WriteString(r.Writer, html)
	return err
}

// Bytes writes binary data
func (r *Response) Bytes(code int, contentType string, data []byte) error {
	r.Writer.Header().Set("Content-Type", contentType)
	r.Writer.WriteHeader(code)
	_, err := r.Writer.Write(data)
	return err
}

// File sends a file as response
func (r *Response) File(filepath string) error {
	http.ServeFile(r.Writer, r.req.HTTPRequest, filepath)
	return nil
}

// Redirect redirects to a URL
func (r *Response) Redirect(code int, url string) error {
	http.Redirect(r.Writer, r.req.HTTPRequest, url, code)
	return nil
}

// Header gets a header value
func (r *Request) Header(key string) string {
	return r.HTTPRequest.Header.Get(key)
}

// SetHeader sets a response header
func (r *Response) SetHeader(key, value string) *Response {
	r.Writer.Header().Set(key, value)
	return r
}

// Method returns the HTTP method
func (r *Request) Method() string {
	return r.HTTPRequest.Method
}

// Path returns the request path
func (r *Request) Path() string {
	return r.HTTPRequest.URL.Path
}

// URL returns the full request URL
func (r *Request) URL() string {
	return r.HTTPRequest.URL.String()
}

// RemoteAddr returns the client IP
func (r *Request) RemoteAddr() string {
	return r.HTTPRequest.RemoteAddr
}

// RequestPool for reusing Request objects
var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{
			params: make(map[string]string, 8),
		}
	},
}

// ResponsePool for reusing Response objects
var responsePool = sync.Pool{
	New: func() interface{} {
		return &Response{}
	},
}

// AcquireRequestResponse gets Request and Response from their pools
func AcquireRequestResponse(w http.ResponseWriter, r *http.Request) (*Request, *Response) {
	req := requestPool.Get().(*Request)
	res := responsePool.Get().(*Response)

	req.HTTPRequest = r
	req.res = res
	req.index = -1

	res.Writer = w
	res.req = req

	return req, res
}

// ReleaseRequestResponse returns objects to their pools
func ReleaseRequestResponse(req *Request, res *Response) {
	req.HTTPRequest = nil
	req.res = nil
	req.params = make(map[string]string, 8)
	req.handlers = nil
	req.index = -1
	requestPool.Put(req)

	res.Writer = nil
	res.req = nil
	responsePool.Put(res)
}
