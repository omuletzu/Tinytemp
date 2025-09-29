package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tinytemp/database"
	"tinytemp/metrics"

	"github.com/go-chi/chi/v5"
)

func AckHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	jobIDStr := chi.URLParam(req, "jobId")
	jobID, _ := strconv.ParseInt(jobIDStr, 10, 64)

	var body map[string]any
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if _, err := database.DB.Exec(ctx, `UPDATE jobs SET status = 'succeeded', updated_at = now(), locked_until = NULL WHERE id = $1`, jobID); err != nil {
		http.Error(w, "cannot set job to succeeded", http.StatusInternalServerError)
		return
	}

	if _, err := database.DB.Exec(ctx, `INSERT INTO jobs_history (job_id, event_type, worker_id, details) VALUES ($1, 'acked', $2, $3)`, jobID, body["worker_id"], body); err != nil {
		http.Error(w, "cannot insert history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	metrics.JobsSucceeded.Inc()
	metrics.JobsInProgress.Dec()

	if body["duration_mils"] != nil {
		metrics.JobsProcessingDuration.Observe(body["duration_mils"].(float64))
	}
}
