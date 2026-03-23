package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brianbeck/c2graph-api/internal/auth"
	"github.com/brianbeck/c2graph-api/internal/config"
	"github.com/brianbeck/c2graph-api/internal/handler"
	neo4jclient "github.com/brianbeck/c2graph-api/internal/neo4j"
	"github.com/brianbeck/c2graph-api/internal/queue"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Pretty console logging for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting C2Graph API Gateway")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to Neo4j
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4jURI, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Neo4j driver")
	}
	defer driver.Close(context.Background())

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Neo4j")
	}
	log.Info().Str("uri", cfg.Neo4jURI).Msg("Connected to Neo4j")

	// Connect to RabbitMQ
	pub, err := queue.NewPublisher(cfg.RabbitMQURI)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer pub.Close()
	log.Info().Msg("Connected to RabbitMQ")

	// Create services
	neo4jClient := neo4jclient.NewClient(driver, cfg.ScanFreshnessHrs)
	authMiddleware := auth.NewMiddleware(cfg.SupabaseJWTSecret)

	// Create handlers
	scanHandler := handler.NewScanHandler(neo4jClient, pub, cfg.MaxWalletsPerScan)
	graphHandler := handler.NewGraphHandler(neo4jClient)
	statusHandler := handler.NewStatusHandler(neo4jClient)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (no auth)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Protected API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authMiddleware.Handler)

		r.Post("/scan", scanHandler.ServeHTTP)
		r.Get("/graph/{address}", graphHandler.ServeHTTP)
		r.Get("/status/{job_id}", statusHandler.ServeHTTP)
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Info().Str("addr", addr).Msg("API Gateway listening")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server failed")
	}
	log.Info().Msg("API Gateway stopped")
}
