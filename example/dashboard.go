package main

import (
	"fmt"
	"log"

	"github.com/nottechdm/notnet/pkg/notnet"
)

func dashboard() {
	app := notnet.New(nil)

	// Initialize stats collector
	notnet.InitStatsCollector()

	// Global middleware
	app.Use(notnet.Logger())
	app.Use(notnet.Recovery())
	app.Use(notnet.CORS(nil))
	app.Use(notnet.Stats()) // Add Stats middleware to track all requests

	// Serve the dashboard HTML
	app.GET("/dashboard", notnet.Dashboard())

	// API endpoint for dashboard data
	app.GET("/api/stats", notnet.StatsAPI())

	app.GET("/", func(req *notnet.Request, res *notnet.Response) error {
		return res.JSON(200, map[string]string{"message": "Welcome to NotNet with Stats Dashboard!"})
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

	// Delayed response to see slow requests
	app.GET("/slow", func(req *notnet.Request, res *notnet.Response) error {
		// Simulate slow processing
		return res.JSON(200, map[string]string{"status": "completed"})
	})

	// Start server
	addr := ":8080"
	fmt.Printf("\n╔════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║     NotNet Server with Stats Dashboard               ║\n")
	fmt.Printf("╠════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║    Server running on http://localhost:8080          ║\n")
	fmt.Printf("║    Dashboard: http://localhost:8080/dashboard       ║\n")
	fmt.Printf("║    API Stats: http://localhost:8080/api/stats       ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════╝\n\n")

	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
