package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tinytemp/database"
	"tinytemp/metrics"

	"github.com/go-chi/chi/v5"
)

func FailHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	jobIdStr := chi.URLParam(req, "jobId")
	jobId, _ := strconv.ParseInt(jobIdStr, 10, 64)

	var body struct {
		WorkerId string `json:"worker_id"`
		Error    string `json:"error"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var attempts int
	var maxAttempts int

	if err := database.DB.QueryRow(ctx, `SELECT attempts, max_attempts FROM jobs WHERE id = $1`, jobId).Scan(&attempts, &maxAttempts); err != nil {
		http.Error(w, "Error while fetching job for failing handler", http.StatusInternalServerError)
		return
	}

	attempts += 1

	if attempts >= maxAttempts {
		if _, err := database.DB.Exec(ctx, `UPDATE jobs SET status = 'DLQ', updated_at = now() WHERE id = $1`, jobId); err != nil {
			http.Error(w, "cannot update to DLQ", http.StatusInternalServerError)
			return
		}
		details := json.RawMessage(fmt.Sprintf(`{"error":"%s","attempts":%d}`, body.Error, attempts))
		if _, err := database.DB.Exec(ctx, `INSERT INTO jobs_history (job_id, event_type, worker_id, details) VALUES ($1, 'DLQ', $2, $3)`, jobId, body.WorkerId, details); err != nil {
			http.Error(w, "cannot insert history", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

		metrics.JobsDLQ.Inc()
		metrics.JobsInProgress.Dec()

		return
	}

	base := 5
	maxDelay := 300
	delay := base << (attempts - 1)
	if delay > maxDelay {
		delay = maxDelay
	}
	jitter := 1 + (attempts % 3)
	delay += jitter

	if _, err := database.DB.Exec(ctx, `UPDATE jobs SET attempts = $1, next_run_at = now() + ($2 || ' seconds')::interval, status = 'pending', locked_until = NULL WHERE id = $3`, attempts, fmt.Sprintf("%d", delay), jobId); err != nil {
		http.Error(w, "cannot update job retry", http.StatusInternalServerError)
		return
	}
	details := json.RawMessage(fmt.Sprintf(`{"error":"%s","attempts":%d}`, body.Error, attempts))
	if _, err := database.DB.Exec(ctx, `INSERT INTO jobs_history (job_id, event_type, worker_id, details) VALUES ($1, 'failed', $2, $3)`, jobId, body.WorkerId, details); err != nil {
		http.Error(w, "cannot insert history", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	metrics.JobsFailed.Inc()
}
