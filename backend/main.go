package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/semanser/ai-coder/assets"
	"github.com/semanser/ai-coder/database"
	"github.com/semanser/ai-coder/executor"
	"github.com/semanser/ai-coder/router"
	"github.com/semanser/ai-coder/services"
)

const defaultPort = "8080"

//go:embed templates/prompts/*.tmpl
var promptTemplates embed.FS

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("failed to read DB env variable")
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("failed to create a pool: %w", err)
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)

	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	defer dbPool.Close()

	db := database.New(dbPool)

	// Setup migrations
	dbMigrationsConnection, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Unable to set dialect: %v\n", err)
	}

	if err := goose.Up(dbMigrationsConnection, "migrations"); err != nil {
		log.Fatalf("Unable to run migrations: %v\n", err)
	}

	log.Println("Migrations ran successfully")

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r := router.New(db)

	assets.Init(promptTemplates)
	services.Init()

	err = executor.InitDockerClient()
	if err != nil {
		log.Fatalf("failed to initialize Docker client: %v", err)
	}

	// Run the server in a separate goroutine
	go func() {
		log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
		if err := http.ListenAndServe(":"+port, r); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-sigChan
	log.Println("Shutting down...")

	// Cleanup resources
	if err := executor.Cleanup(db); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	log.Println("Shutdown complete")
}
