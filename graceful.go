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

// Shutdown can gracefully shut down a Go application and run a function for additional cleanup tasks.
func Shutdown(shutdownFunc, cleanupFunc func(), gracePeriod time.Duration) {
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
			log.Println("Application context is done")
			return
		}
	}()

	// Wait for an interrupt signal (e.g. SIGINT or SIGTERM)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-signalCh

	// Signal the context to cancel
	cancel()

	// Set a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), gracePeriod)
	defer shutdownCancel()

	// Run the shutdown function, if provided
	if shutdownFunc != nil {
		log.Println("Run the shutdown function...")
		shutdownFuncWithTimeout(shutdownCtx, shutdownFunc)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Run the cleanup function, if provided
	if cleanupFunc != nil {
		log.Println("Run the cleanup function...")
		cleanupFunc()
	}

	log.Println("Application stopped gracefully")
}

func shutdownFuncWithTimeout(ctx context.Context, shutdownFunc func()) {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		shutdownFunc()
	}()
	select {
	case <-doneCh:
		// Shutdown function completed successfully
		log.Println("Shutdown function completed successfully")
	case <-ctx.Done():
		// Shutdown function took too long to complete
		log.Println("Shutdown function timed out")
	}
}
