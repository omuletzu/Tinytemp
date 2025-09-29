package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"tinytemp/database"
	"tinytemp/handlers"
	"tinytemp/metrics"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = godotenv.Load()
	if err := database.InitDB(context.Background(), os.Getenv("DB_URL")); err != nil {
		log.Fatalf("DB init: %v", err)

	}

	defer database.DB.Close()

	r := chi.NewRouter()

	r.Post("/enqueue", handlers.EnqueueHandler)
	r.Post("/next-job", handlers.NextJobHandler)
	r.Post("/ack/{jobId}", handlers.AckHandler)
	r.Post("/fail/{jobId}", handlers.FailHandler)
	r.Post("/heartbeat/{jobId}", handlers.HeartbeatHandler)

	metrics.InitProm()
	r.Handle("/metrics", promhttp.Handler())

	serverPort := 8000

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverPort),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Cannot start server")
	}

	log.Println("Server started")
}
