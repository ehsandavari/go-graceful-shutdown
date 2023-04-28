package graceful

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Graceful can gracefully shut down a Go application and run a function for additional cleanup tasks.
func Graceful(shutdownFunc, cleanupFunc func(), gracePeriod time.Duration) {
	// Use a WaitGroup to track active goroutines
	var wg sync.WaitGroup

	// Create a context that will be canceled when the interrupt signal is received
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Application started")
		select {
		case <-ctx.Done():
			log.Println("Application stopped gracefully")
			return
		}
	}()

	// Wait for an interrupt signal (e.g. SIGINT or SIGTERM)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-signalCh

	// Signal the context to cancel
	cancel()

	// Run the shutdown function, if provided
	if shutdownFunc != nil {
		log.Println("Run the shutdown function...")
		shutdownFunc()
	}

	// Set a deadline for shutdown
	ctx, timeoutCancel := context.WithTimeout(ctx, gracePeriod)
	defer timeoutCancel()

	// Wait for all goroutines to finish
	wg.Wait()

	// Run the cleanup function, if provided
	if cleanupFunc != nil {
		log.Println("Run the cleanup function...")
		cleanupFunc()
	}

	log.Println("Application stopped gracefully")
}
