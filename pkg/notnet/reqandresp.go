package notnet

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Response represents an HTTP response wrapper
//
// It provides methods for sending different types of responses (JSON, HTML, files, etc.) and setting headers
// The Response struct contains a reference to the http.ResponseWriter and the associated Request, allowing it to access request data and write responses
type Response struct {
	// Writer is the http.ResponseWriter used to send the response back to the client
	Writer http.ResponseWriter
	// req is a reference to the associated Request, allowing access to request data and context
	req *Request
}

// Request represents an HTTP request wrapper
//
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
//
// It takes a key and returns the corresponding path parameter value from the request's params map
func (r *Request) Param(key string) string {
	return r.params[key]
}

// Query returns a query parameter value
//
// It takes a key and returns the corresponding query parameter value from the request's URL query parameters
func (r *Request) Query(key string) string {
	return r.HTTPRequest.URL.Query().Get(key)
}

// Next moves to the next handler in the chain
//
// It increments the handler index and calls the next handler function in the handlers slice if there are more handlers to execute
func (r *Request) Next() error {
	r.index++
	if r.index < len(r.handlers) {
		return r.handlers[r.index](r, r.res)
	}
	return nil
}

// Status sets the HTTP status code
//
// It takes a status code and writes it to the response header, allowing the user to set the desired HTTP status for the response
func (r *Response) Status(code int) *Response {
	r.Writer.WriteHeader(code)
	return r
}

// String writes a string response
//
// It takes a status code and a string, sets the Content-Type header to text/plain, writes the status code, and sends the string as the response body
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
//
// It takes a status code and an arbitrary data structure, sets the Content-Type header to application/json, writes the status code, and encodes the data as JSON in the response body
func (r *Response) JSON(code int, data interface{}) error {
	r.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Writer.WriteHeader(code)
	return json.NewEncoder(r.Writer).Encode(data)
}

// BindJSON parses the request body as JSON
//
// It takes a pointer to a struct and decodes the JSON request body into that struct, allowing users to easily parse incoming JSON data into Go structs
func (r *Request) BindJSON(v interface{}) error {
	return json.NewDecoder(r.HTTPRequest.Body).Decode(v)
}

// HTML writes an HTML response
//
// It takes a status code and an HTML string, sets the Content-Type header to text/html, writes the status code, and sends the HTML string as the response body
func (r *Response) HTML(code int, html string) error {
	r.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Writer.WriteHeader(code)
	_, err := io.WriteString(r.Writer, html)
	return err
}

// Bytes writes binary data
//
// It takes a status code, a content type, and a byte slice, sets the Content-Type header to the specified content type, writes the status code, and sends the byte slice as the response body
func (r *Response) Bytes(code int, contentType string, data []byte) error {
	r.Writer.Header().Set("Content-Type", contentType)
	r.Writer.WriteHeader(code)
	_, err := r.Writer.Write(data)
	return err
}

// File sends a file as response
//
// It takes a file path and uses http.ServeFile to send the file as the response, allowing users to easily serve static files or downloads
func (r *Response) File(filepath string) error {
	http.ServeFile(r.Writer, r.req.HTTPRequest, filepath)
	return nil
}

// Redirect redirects to a URL
//
// It takes a status code and a URL, and uses http.Redirect to send a redirect response to the client, allowing users to easily redirect clients to different URLs
func (r *Response) Redirect(code int, url string) error {
	http.Redirect(r.Writer, r.req.HTTPRequest, url, code)
	return nil
}

// Header gets a header value
//
// It takes a header key and returns the corresponding header value from the request's HTTP headers, allowing users to access incoming request headers easily
func (r *Request) Header(key string) string {
	return r.HTTPRequest.Header.Get(key)
}

// SetHeader sets a response header
//
// It takes a header key and value, and sets that header in the response using the http.ResponseWriter's Header().Set method, allowing users to easily set custom headers in their responses
func (r *Response) SetHeader(key, value string) *Response {
	r.Writer.Header().Set(key, value)
	return r
}

// SSE sets up the response for Server-Sent Events
//
// It sets the Content-Type to text/event-stream and other necessary headers for SSE, and flushes the initial headers to the client
func (r *Response) SSE() *Response {
	r.Writer.Header().Set("Content-Type", "text/event-stream")
	r.Writer.Header().Set("Cache-Control", "no-cache")
	r.Writer.Header().Set("Connection", "keep-alive")
	r.Writer.Header().Set("Transfer-Encoding", "chunked")
	r.Writer.WriteHeader(http.StatusOK)
	if f, ok := r.Writer.(http.Flusher); ok {
		f.Flush()
	}
	return r
}

// SendEvent sends a server-sent event to the client
//
// It formats the data according to the SSE protocol (event: <name>\ndata: <data>\n\n).
// If data is not a string or byte slice, it will be encoded as JSON. It also flushes the event to the client immediately.
func (r *Response) SendEvent(event string, data interface{}) error {
	if event != "" {
		if _, err := fmt.Fprintf(r.Writer, "event: %s\n", event); err != nil {
			return err
		}
	}

	var d []byte
	var err error
	switch v := data.(type) {
	case string:
		d = []byte(v)
	case []byte:
		d = v
	default:
		d, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(r.Writer, "data: %s\n\n", string(d)); err != nil {
		return err
	}

	if f, ok := r.Writer.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// Method returns the HTTP method
//
// It returns the HTTP method of the incoming request by accessing the Method field of the original http.Request, allowing users to easily determine the request method in their handlers
func (r *Request) Method() string {
	return r.HTTPRequest.Method
}

// Path returns the request path
//
// It returns the URL path of the incoming request by accessing the Path field of the original http.Request's URL, allowing users to easily access the request path in their handlers
func (r *Request) Path() string {
	return r.HTTPRequest.URL.Path
}

// URL returns the full request URL
//
// It returns the full URL of the incoming request by calling the String() method on the original http.Request's URL, allowing users to easily access the complete request URL in their handlers
func (r *Request) URL() string {
	return r.HTTPRequest.URL.String()
}

// RemoteAddr returns the client IP
//
// It returns the remote address of the client making the request by accessing the RemoteAddr field of the original http.Request, allowing users to easily access the client's IP address in their handlers
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
//
// It retrieves a Request and Response object from their respective sync.Pool instances, sets up the necessary references between them, and returns them for use in handling an incoming HTTP request
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
//
// It takes a Request and Response object, resets their fields to clear any request-specific data, and puts them back into their respective sync.Pool instances for reuse in future requests, helping to reduce memory allocations and improve performance
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
