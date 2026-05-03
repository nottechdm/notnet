package main

import (
	"fmt"
	"time"

	"github.com/nottechdm/notnet/pkg/notnet"
)

func sse() {
	app := notnet.New(nil)

	app.GET("/events", func(req *notnet.Request, res *notnet.Response) error {
		// Set up SSE
		res.SSE()

		// Send events every 2 seconds
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// Get the request context to detect client disconnection
		ctx := req.HTTPRequest.Context()

		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("Client disconnected")
				return nil
			case t := <-ticker.C:
				msg := fmt.Sprintf("Message %d at %s", i, t.Format("15:04:05"))
				fmt.Printf("Sending: %s\n", msg)

				// Send a structured event
				err := res.SendEvent("message", map[string]interface{}{
					"id":      i,
					"content": msg,
					"time":    t.Unix(),
				})
				if err != nil {
					return err
				}
			}
		}

		// Send a final event
		res.SendEvent("close", "Connection closing")
		return nil
	})

	fmt.Println("Server starting on :8080...")
	fmt.Println("Open http://localhost:8080/events in your browser or use: curl -N http://localhost:8080/events")
	if err := app.Listen(":8080"); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
