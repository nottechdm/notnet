# NotNet
A lightweight HTTP router written in Go.

![Release](https://img.shields.io/github/v/release/nottechdm/notnet)
![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)
[![GoDoc](https://pkg.go.dev/badge/github.com/nottechdm/notnet.svg)](https://pkg.go.dev/github.com/nottechdm/notnet)
![Tests](https://github.com/nottechdm/notnet/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/nottechdm/notnet)](https://goreportcard.com/report/github.com/nottechdm/notnet)
[![codecov](https://codecov.io/gh/nottechdm/notnet/branch/main/graph/badge.svg)](https://codecov.io/gh/nottechdm/notnet)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

**NotNet** is a lightweight, high-performance, and ergonomic routing framework for Go. It allows you to quickly structure RESTful HTTP services with an expressive API, built-in middleware, and custom group routing.

<img width="1983" height="793" alt="ChatGPT Image Apr 23, 2026, 08_44_43 PM" src="https://github.com/user-attachments/assets/ffdcf831-d06e-43ff-81f1-b0f106d82e4f" />


## Features

- **Expressive Routing**: Simple `.GET()`, `.POST()`, `.PUT()`, etc., for robust route handling.
- **Route Groups**: Group endpoints with shared prefixes and nested middleware stacks.
- **Path Parameters**: Extract dynamic URL parameters like `/users/:id` out of the box.
- **Pre-packaged Middleware**:
  - `Logger`: Detailed request logging.
  - `Recovery`: Auto-recovery from application panics.
  - `CORS`: Simplifies Cross-Origin Resource Sharing.
  - `RateLimit`: Prevents abuse through IP-based request limits.
  - `AuthRequired`: Simple stubbed bearer token checking.
  - `RequestID`: Track requests universally with trace IDs.
- **Extensible Request/Response Engine**: Bind JSON payloads quickly, return robust JSON, HTML or strings using dedicated `Response/Request` wrappers.

## Installation

Run this in your Go module project to get `notnet`:

```bash
go get github.com/nottechdm/notnet
```

## Quick Start

Creating an API in `NotNet` takes only a few lines:

```go
package main

import (
	"log"
	"github.com/nottechdm/notnet/pkg/notnet"
)

func main() {
	app := notnet.New(nil)

	// Attach Universal Middleware
	app.Use(notnet.Logger())
	app.Use(notnet.Recovery())

	// Simple Route
	app.GET("/ping", func(req *notnet.Request, res *notnet.Response) error {
		return res.String(200, "pong")
	})

	// Dynamic Path Parameters
	app.GET("/users/:id", func(req *notnet.Request, res *notnet.Response) error {
		id := req.Param("id")
		return res.JSON(200, map[string]string{"user_id": id})
	})

	// Read JSON Payload
	app.POST("/api/data", func(req *notnet.Request, res *notnet.Response) error {
		var payload map[string]interface{}
		if err := req.BindJSON(&payload); err != nil {
			return res.JSON(400, map[string]string{"error": "invalid json"})
		}
		
		payload["received"] = true
		return res.JSON(201, payload)
	})

	log.Println("Server running on :8080")
	log.Fatal(app.Listen(":8080"))
}
```

## Route Groups & Custom Middleware

You can isolate your authentication handlers from public endpoints easily by using groups:

```go
// Public
app.GET("/status", func(req *notnet.Request, res *notnet.Response) error {
	return res.JSON(200, map[string]string{"status": "ok"})
})

// Protected API V1 Group 
api := app.Group("/api/v1", notnet.AuthRequired())

api.GET("/dashboard", func(req *notnet.Request, res *notnet.Response) error {
	return res.JSON(200, map[string]string{"msg": "You have access!"})
})
```

## Custom Handlers

`NotNet` comes with default 404, error, and panic handling, but you can override them with your unique application responses at any time:

```go
app.SetNotFoundHandler(func(req *notnet.Request, res *notnet.Response) {
    res.JSON(404, map[string]string{
        "error": "This page does not exist",
        "path":  req.Path(),
    })
})
```

## License

This project is licensed under the MIT License.
