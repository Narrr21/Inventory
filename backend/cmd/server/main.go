package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"my-backend/internal/db"
	"my-backend/internal/handlers"
	"my-backend/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on already-set environment variables")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	client, err := db.Connect(connectCtx, db.LoadConfig())
	cancel()
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.Disconnect(disconnectCtx, client); err != nil {
			log.Printf("error disconnecting from MongoDB: %v", err)
		}
	}()

	itemRepo := repository.NewItemRepository(client.Database)
	projectRepo := repository.NewProjectRepository(client.Database)
	itemTypeRepo := repository.NewItemTypeRepository(client.Database)
	itemHandler := handlers.NewItemHandler(itemRepo, projectRepo, itemTypeRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo, itemRepo)
	itemTypeHandler := handlers.NewItemTypeHandler(itemTypeRepo, itemRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(itemRepo, projectRepo, itemTypeRepo)
	router := handlers.NewRouter(itemHandler, projectHandler, itemTypeHandler, analyticsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("API listening on port %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
}
