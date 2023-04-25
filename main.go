package main

import (
	"context"
	"github.com/ehsandavari/golang-graceful-shutdown/graceful"
	"github.com/ehsandavari/golang-graceful-shutdown/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"time"
)

// Example HTTP server using Gin
func startHTTPServer() *http.Server {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		time.Sleep(4 * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, world!",
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}

// Example gRPC server
type server struct {
	proto.UnimplementedGreeterServer
}

func (*server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloResponse, error) {
	time.Sleep(4 * time.Second)
	return &proto.HelloResponse{Message: "Hello, " + in.GetName() + "!"}, nil
}

func startGRPCServer() *grpc.Server {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterGreeterServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return s
}

func main() {
	// Start HTTP server
	httpServer := startHTTPServer()

	// Start gRPC server
	grpcServer := startGRPCServer()

	// Define shutdown function to gracefully stop servers
	shutdownFunc := func() {
		log.Println("Shutting down gracefully...")
		grpcServer.GracefulStop()
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server shutdown error: %v\n", err)
		}
	}

	// Define cleanup function to close database connections or do any other necessary cleanup
	cleanupFunc := func() {
		log.Println("Cleaning up...")
	}
	// Call Graceful function with cleanup and shutdown functions
	graceful.Graceful(shutdownFunc, cleanupFunc, 10*time.Second)
}
