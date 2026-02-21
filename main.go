package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neocomx/TaskQueueService/api"
	"github.com/neocomx/TaskQueueService/task"
	"github.com/neocomx/TaskQueueService/worker"
)

func main() {
	store := task.NewStore()
	processor := &worker.PrintProcessor{}
	pool := worker.NewPool(3, store, processor, 5*time.Second)

	pool.Start()

	server := api.NewServer(store, pool)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Println("Server listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-quit
	fmt.Println("\n Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("HTTP server shutdown error: %v \n", err)
	}

	pool.Shutdown()
}
