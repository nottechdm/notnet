package main

import (
	"time"

	"github.com/nottechdm/notnet/pkg/notnet"
)

func ApplyConfig() {
	// Create a new NotNet application with custom engine options
	app := notnet.New(&notnet.EngineOption{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	})

	app.Use(notnet.Logger())

	// This route will use the default engine timeouts
	app.GET("/default-timeout", func(r1 *notnet.Request, r2 *notnet.Response) error {
		r2.HTML(200, "<html>Hello NotNet</html>")
		return nil
	})

	// This route will override the engine timeouts with custom values
	app.GET("/long-timeout", func(r1 *notnet.Request, r2 *notnet.Response) error {
		r2.HTML(200, "<html>Hello NotNet</html>")
		return nil
	}).ApplyConfig(&notnet.EngineOption{
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 40 * time.Second,
	})
}
