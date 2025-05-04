package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/XEDJK/ToT/db/database" // Import generated db code
	"github.com/XEDJK/ToT/handlers"    // Import handlers
	"github.com/XEDJK/ToT/middleware"  // Import middleware
	"github.com/XEDJK/ToT/routes"      // Import routes
	"github.com/jackc/pgx/v5/pgxpool"  // Import pgx driver
	"github.com/joho/godotenv"         // Import godotenv for loading environment variables
)

func main() {

	const filepathRoot = "."

	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("$PORT must be set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("$DB_URL must be set")
	}

	conn, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Cant't connect to the database: ", err)
	}

	// Ping the database to verify connection
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	db := database.New(conn)

	defer func() {
		conn.Close()
	}()

	// Instantiate the APIConfig from handlers package
	apiCfg := handlers.NewAPIConfig(db)

	mux := http.NewServeMux()

	// Use the handler method from apiCfg
	routes.RegisterRoutes(mux, apiCfg)

	// Apply logging middleware
	handlerWithCors := middleware.CorsMiddleware(mux)
	loggedHandler := middleware.LoggingMiddleware(handlerWithCors)

	srv := &http.Server{
		Addr:         ":" + portString,
		Handler:      loggedHandler,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting server on port " + portString)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
