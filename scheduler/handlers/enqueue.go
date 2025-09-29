package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"tinytemp/database"
	"tinytemp/metrics"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EnqueueRequest struct {
	JobType        string          `json:"job_type"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	MaxAttempts    *int            `json:"max_attempts,omitempty"`
}

func EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req EnqueueRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "cannot decode body", http.StatusBadRequest)
		return
	}
	if req.JobType == "" {
		http.Error(w, "job_type required", http.StatusBadRequest)
		return
	}
	maxAttempts := 5
	if req.MaxAttempts != nil {
		maxAttempts = *req.MaxAttempts
	}

	if req.IdempotencyKey != nil {
		var existingID int64
		err := database.DB.QueryRow(ctx, `SELECT id FROM jobs WHERE idempotency_key = $1 LIMIT 1`, *req.IdempotencyKey).Scan(&existingID)
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"job_id": existingID, "existing": true})
			return
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	var jobId int64
	err := database.DB.QueryRow(ctx, `
        INSERT INTO jobs (job_type, payload, idempotency_key, max_attempts)
        VALUES ($1, $2, $3, $4) RETURNING id
    `, req.JobType, req.Payload, req.IdempotencyKey, maxAttempts).Scan(&jobId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var existingID int64
			if e := database.DB.QueryRow(ctx, `SELECT id FROM jobs WHERE idempotency_key = $1 LIMIT 1`, req.IdempotencyKey).Scan(&existingID); e == nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"job_id": existingID, "existing": true})
				return
			}
		}
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	_, _ = database.DB.Exec(ctx, `INSERT INTO jobs_history (job_id, event_type, details) VALUES ($1, 'enqueued', $2)`, jobId, json.RawMessage(`{}`))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"job_id": jobId})

	metrics.JobsTotal.Inc()
}
