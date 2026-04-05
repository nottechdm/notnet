package main

import (
	"fmt"
	"log"

	"github.com/nottechdm/notnet/pkg/notnet"
)

func main() {
	app := notnet.New(nil)

	// Global middleware
	app.Use(notnet.Logger())
	app.Use(notnet.Recovery())
	app.Use(notnet.CORS(nil))

	// Basic routes
	app.GET("/", func(req *notnet.Request, res *notnet.Response) error {
		return res.JSON(200, map[string]string{"message": "Hello from Apex!"})
	})

	app.GET("/ping", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "pong")
	})

	// Path parameters
	app.GET("/users/:id", func(req *notnet.Request, res *notnet.Response) error {
		id := req.Param("id")
		return res.JSON(200, map[string]string{"user_id": id})
	})

	app.GET("/posts/:id/comments/:comment_id", func(req *notnet.Request, res *notnet.Response) error {
		postID := req.Param("id")
		commentID := req.Param("comment_id")
		return res.JSON(200, map[string]string{
			"post_id":    postID,
			"comment_id": commentID,
		})
	})

	// POST with JSON
	app.POST("/api/data", func(req *notnet.Request, res *notnet.Response) error {
		var payload map[string]interface{}
		if err := req.BindJSON(&payload); err != nil {
			return res.JSON(400, map[string]string{"error": "invalid json"})
		}
		payload["received"] = true
		return res.JSON(201, payload)
	})

	// Query parameters
	app.GET("/search", func(req *notnet.Request, res *notnet.Response) error {
		q := req.Query("q")
		if q == "" {
			return res.JSON(400, map[string]string{"error": "missing query parameter"})
		}
		return res.JSON(200, map[string]string{"query": q})
	})

	// Multiple query params
	app.GET("/filter", func(req *notnet.Request, res *notnet.Response) error {
		category := req.Query("category")
		sort := req.Query("sort")
		return res.JSON(200, map[string]string{
			"category": category,
			"sort":     sort,
		})
	})

	// Route group with auth middleware
	api := app.Group("/api/v1", notnet.AuthRequired())
	api.GET("/status", func(req *notnet.Request, res *notnet.Response) error {
		return res.JSON(200, map[string]string{"status": "ok"})
	})

	api.POST("/protected", func(req *notnet.Request, res *notnet.Response) error {
		return res.JSON(200, map[string]string{"protected": "data"})
	})

	api.DELETE("/resource/:id", func(req *notnet.Request, res *notnet.Response) error {
		id := req.Param("id")
		return res.JSON(200, map[string]string{"deleted": id})
	})

	// Redirect example
	app.GET("/old", func(req *notnet.Request, res *notnet.Response) error {
		return res.Redirect(301, "/new")
	})

	app.GET("/new", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "New route!")
	})

	// HTML response
	app.GET("/html", func(req *notnet.Request, res *notnet.Response) error {
		return res.HTML(200, "<h1>Hello HTML</h1>")
	})

	// Custom error handling
	app.SetErrorHandler(func(req *notnet.Request, res *notnet.Response, err error) {
		fmt.Printf("Error occurred: %v\n", err)
		res.JSON(500, map[string]string{"error": err.Error()})
	})

	// Custom 404
	app.SetNotFoundHandler(func(req *notnet.Request, res *notnet.Response) {
		res.JSON(404, map[string]string{
			"error": "endpoint not found",
			"path":  req.Path(),
		})
	})

	// Start server
	log.Println("NotNet server starting on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
