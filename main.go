package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/froggu-tantei/ToT/db/database" // Import generated db code
	"github.com/froggu-tantei/ToT/handlers"    // Import handlers
	"github.com/froggu-tantei/ToT/middleware"  // Import middleware
	"github.com/froggu-tantei/ToT/routes"      // Import routes
	"github.com/froggu-tantei/ToT/storage"     // Import storage
	"github.com/jackc/pgx/v5/pgxpool"          // Import pgx driver
	"github.com/joho/godotenv"                 // Import godotenv for loading environment variables
)

func main() {

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
		log.Fatal("Can't connect to the database: ", err)
	}

	// Ping the database to verify connection
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	db := database.New(conn)

	// Rate limiting configuration with fallbacks
	authLimit := getEnvAsInt("AUTH_RATE_LIMIT", 3)          // Default: 3 requests
	authWindow := getEnvAsInt("AUTH_RATE_WINDOW", 60)       // Default: 60 seconds
	genericLimit := getEnvAsInt("GENERIC_RATE_LIMIT", 30)   // Default: 30 requests
	genericWindow := getEnvAsInt("GENERIC_RATE_WINDOW", 60) // Default: 60 seconds

	// Convert to rate (requests per second) and create configs
	authRate := float64(authLimit) / float64(authWindow)
	genericRate := float64(genericLimit) / float64(genericWindow)

	// Create rate limiter configs
	authConfig := middleware.RateLimiterConfig{
		Rate:            authRate,
		Capacity:        authLimit,
		MaxBuckets:      10000,
		CleanupInterval: 5 * time.Minute,
		BucketTTL:       10 * time.Minute,
		MaxRetryAfter:   5 * time.Minute,
	}

	genericConfig := middleware.RateLimiterConfig{
		Rate:            genericRate,
		Capacity:        genericLimit,
		MaxBuckets:      10000,
		CleanupInterval: 5 * time.Minute,
		BucketTTL:       10 * time.Minute,
		MaxRetryAfter:   5 * time.Minute,
	}

	// Create rate limiters with proper configs
	authLimiter := middleware.NewRateLimiter(authConfig)
	genericLimiter := middleware.NewRateLimiter(genericConfig)

	// Ensure proper cleanup on shutdown
	defer func() {
		if err := authLimiter.Close(); err != nil {
			log.Printf("Error closing auth limiter: %v", err)
		}
		if err := genericLimiter.Close(); err != nil {
			log.Printf("Error closing generic limiter: %v", err)
		}
	}()

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

	defer conn.Close()
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

// Helper function to get environment variable as int with fallback
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid value for %s: %s, using fallback: %d", key, value, fallback)
	}
	return fallback
}
