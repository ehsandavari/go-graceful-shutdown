package graceful

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestGracefulShutdown(t *testing.T) {
	// Set up a mock application
	var shutdownCalled bool
	var cleanupCalled bool
	var mockAppStarted bool
	var mockAppExited bool

	shutdownFunc := func() {
		shutdownCalled = true
	}

	cleanupFunc := func() {
		cleanupCalled = true
	}

	mockApp := func() {
		mockAppStarted = true
		select {
		case <-time.After(time.Second):
			mockAppExited = true
		}
	}

	go func() {
		Shutdown(shutdownFunc, cleanupFunc, 10*time.Second)
	}()

	// Wait for a short period of time before starting the mock application
	// to make sure the Shutdown function is running before the application starts
	time.Sleep(time.Millisecond * 500)

	go func() {
		mockApp()
	}()

	// Wait for the mock application to start
	time.Sleep(time.Millisecond * 500)

	// Send an interrupt signal to the process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		return
	}

	// Wait for the application to shut down gracefully
	time.Sleep(500 * time.Millisecond)

	// Check that the application started and exited
	if !mockAppStarted {
		t.Errorf("Mock application did not start")
	}
	if !mockAppExited {
		t.Errorf("Mock application did not exit")
	}

	// Check that the shutdown and cleanup functions were called
	if !shutdownCalled {
		t.Errorf("Shutdown function was not called")
	}
	if !cleanupCalled {
		t.Errorf("Cleanup function was not called")
	}
}

func TestShutdownFuncWithTimeout(t *testing.T) {
	// Create a context with a short deadline
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create a mock shutdown function
	var shutdownCalled bool
	shutdownFunc := func() {
		shutdownCalled = true
		time.Sleep(200 * time.Millisecond)
	}

	// Call the shutdownFuncWithTimeout function with the mock shutdown function
	shutdownFuncWithTimeout(ctx, shutdownFunc)

	// Check that the shutdown function was called
	if !shutdownCalled {
		t.Error("Shutdown function was not called")
	}
}
