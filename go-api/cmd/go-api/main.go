// Command go-api is the real CRUD server for this tutorial. It's
// deliberately thin: parse flags, build the router (internal/api), start an
// http.Server, and shut it down gracefully on interrupt — everything else
// lives in the internal/ packages it composes.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api/internal/api"
)

func main() {
	port := flag.String("port", "3005", "port to listen on")
	flag.Parse()

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: api.NewRouter(),
	}

	// Start the server on its own goroutine so main() is free to block on
	// the shutdown signal below instead of on ListenAndServe itself (see
	// docs/go/02-concurrency.md for goroutine basics).
	go func() {
		log.Printf("Go server running on http://localhost:%s\n", *port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until we receive SIGINT (Ctrl-C) or SIGTERM (e.g. `kill`, or a
	// container/orchestrator asking the process to stop). This only works
	// reliably when running the compiled binary directly — see
	// docs/go/04-api.md for why `go run .` complicates signal delivery.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")

	// Give in-flight requests up to 5 seconds to finish before forcing the
	// server closed. context.WithTimeout + select-style blocking (here via
	// Shutdown itself, which internally waits on the context) is the
	// standard Go pattern for "wait, but not forever" (see
	// docs/go/04-api.md).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("Server stopped.")
}
