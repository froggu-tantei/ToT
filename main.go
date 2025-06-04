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
	"github.com/XEDJK/ToT/storage"     // Import storage
	"github.com/jackc/pgx/v5/pgxpool"  // Import pgx driver
	"github.com/joho/godotenv"         // Import godotenv for loading environment variables
)

func main() {

	const filepathRoot = "."

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

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

	authLimiter := middleware.NewRateLimiter(5, 10)     // 5 requests per 10 seconds
	genericLimiter := middleware.NewRateLimiter(20, 60) // 20 requests per 60 seconds

	fileStorage := storage.NewLocalStorage("uploads", "")
	// Change fileStorage into this whenever I want to use S3 storage:
	// fileStorage, err := storage.NewS3Storage(
	// "your-bucket-name",
	// "your-region",  // e.g., "eu-west-1"
	// ""  // Optional CDN URL if I have one
	// )
	// if err != nil {
	// log.Fatal("Failed to initialize S3 storage:", err)
	//}

	// Instantiate the APIConfig from handlers package
	apiCfg := handlers.NewAPIConfig(db, fileStorage)

	// Create Chi router (this handles all middleware internally)
	router := routes.RegisterRoutes(apiCfg, authLimiter, genericLimiter)

	// Serve static files using Chi.
	router.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	srv := &http.Server{
		Addr:         ":" + portString,
		Handler:      router,
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
