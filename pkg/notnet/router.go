package notnet

import (
	"strings"
)

// Route represents a registered route
type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

// Router is the route matcher
type Router struct {
	tree map[string]*RadixNode
}

// RadixNode is a node in the radix tree
type RadixNode struct {
	edge       string
	handler    HandlerFunc
	param      string                // parameter name if this is a param node
	isParam    bool                  // true if this node represents a parameter
	children   map[string]*RadixNode // exact match children
	paramChild *RadixNode            // wildcard/param child
	method     string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		tree: make(map[string]*RadixNode),
	}
}

// Register registers a route
func (r *Router) Register(method, path string, handler HandlerFunc) {
	key := method + ":" + path

	if r.tree[key] == nil {
		r.tree[key] = &RadixNode{
			edge:     path,
			handler:  handler,
			method:   method,
			children: make(map[string]*RadixNode),
		}
	} else {
		r.tree[key].handler = handler
	}
}

// Match matches a request path against registered routes
func (r *Router) Match(method, path string) (*Route, map[string]string, bool) {
	params := make(map[string]string)

	// Exact lookup first for performance
	key := method + ":" + path
	if node, ok := r.tree[key]; ok && node.handler != nil {
		return &Route{Method: method, Path: path, Handler: node.handler}, params, true
	}

	// Try to match with parameters
	for _, node := range r.tree {
		if node.method != method {
			continue
		}

		if matched, p := matchPath(node.edge, path); matched {
			params = p
			if node.handler != nil {
				return &Route{Method: method, Path: node.edge, Handler: node.handler}, params, true
			}
		}
	}

	return nil, nil, false
}

// matchPath matches a path pattern against a request path
func matchPath(pattern, path string) (bool, map[string]string) {
	params := make(map[string]string)

	parts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(parts) != len(pathParts) {
		return false, params
	}

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			// Parameter match
			paramName := part[1:]
			params[paramName] = pathParts[i]
		} else if part != pathParts[i] {
			// Exact match failed
			return false, params
		}
	}

	return true, params
}
