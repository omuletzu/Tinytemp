package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"
	"tinytemp/database"
	"tinytemp/metrics"
	"tinytemp/models"
)

type NextJobResponse struct {
	Jobs []models.Job `json:"jobs"`
}

type JobPriority struct {
	Job      models.Job
	Priority float64
}

type WorkerBodyRequest struct {
	WorkerCPU int `json:"WorkerCPU"`
	WorkerGPU int `json:"WorkerGPU"`
	WorkerMEM int `json:"WorkerMEM"`
}

func get_predicted_runtime(payload json.RawMessage) float64 {
	defaultErrorReturn := 60.0

	url := "http://localhost:8001/predict"

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return defaultErrorReturn
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return defaultErrorReturn
	}

	var result map[string]float64

	if err := json.Unmarshal(body, &result); err != nil {
		return defaultErrorReturn
	}

	return result["runtime"]
}

func NextJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	q := req.URL.Query()
	workerId := q.Get("worker_id")
	capacity := 1

	cptStr := q.Get("capacity")

	streamBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Cannot read worker body stream", http.StatusInternalServerError)
		return
	}

	var workerBody WorkerBodyRequest
	if err := json.Unmarshal(streamBody, &workerBody); err != nil {
		http.Error(w, "Cannot read worker body from stream", http.StatusInternalServerError)
		return
	}

	if cptStr != "" {
		if v, err := strconv.Atoi(cptStr); err == nil && v > 0 {
			capacity = v
		}
	}

	leaseTime := 30

	jobs := make([]models.Job, 0, capacity)

	rows, err := database.DB.Query(ctx, `
		SELECT id, payload, created_at
		FROM jobs
		WHERE status = 'pending' AND next_run_at <= now() AND attempts < max_attempts
		LIMIT 100
	`)

	if err != nil {
		http.Error(w, "Cannot fetch initial job list from DB", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	jobsPriority := []JobPriority{}
	now := time.Now()
	alpha := 0.01

	for rows.Next() {
		var job models.Job

		if err := rows.Scan(&job.Id, &job.Payload, &job.CreatedAt); err != nil {
			continue
		}

		var payloadMap map[string]interface{}
		if err := json.Unmarshal(job.Payload, &payloadMap); err != nil {
			continue
		}

		payloadMap["worker_cpu"] = workerBody.WorkerCPU
		payloadMap["worker_gpu"] = workerBody.WorkerGPU
		payloadMap["worker_mem"] = workerBody.WorkerMEM

		finalPayload, err := json.Marshal(payloadMap)

		if err != nil {
			continue
		}

		predictedRuntime := get_predicted_runtime(finalPayload)

		priority := predictedRuntime / (1 + alpha*now.Sub(job.CreatedAt).Seconds())

		jobsPriority = append(jobsPriority, JobPriority{
			Job:      job,
			Priority: priority,
		})
	}

	if len(jobsPriority) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sort.Slice(jobsPriority, func(i, j int) bool {
		return jobsPriority[i].Priority < jobsPriority[j].Priority
	})

	jobsPriorityIdsOnly := make([]int64, len(jobsPriority))
	for i, jp := range jobsPriority {
		jobsPriorityIdsOnly[i] = jp.Job.Id
	}

	txCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := database.DB.Begin(txCtx)
	if err != nil {
		http.Error(w, "Cannot start transcation with DB", http.StatusInternalServerError)
		return
	}

	jobsClaimedRows, err := tx.Query(txCtx, `
		WITH top_jobs AS (
			SELECT id
			FROM jobs
			WHERE id = ANY($2) AND status = 'pending'
			ORDER BY array_position($2::bigint[], id)
			FOR UPDATE SKIP LOCKED
			LIMIT $3
		)
		UPDATE jobs
		SET status = 'in_progress',
			locked_until = now() + ($1::int || ' seconds')::interval,
			attempts = attempts + 1,
			updated_at = now()
		WHERE id in (SELECT id FROM top_jobs)
		RETURNING id, payload, created_at
	`, fmt.Sprintf("%d", leaseTime), jobsPriorityIdsOnly, capacity)

	if err != nil {
		tx.Rollback(txCtx)
		http.Error(w, "Cannot claim jobs", http.StatusInternalServerError)
		return
	}

	jobsClaimed := []models.Job{}
	for jobsClaimedRows.Next() {
		var j models.Job
		if err := jobsClaimedRows.Scan(&j.Id, &j.Payload, &j.CreatedAt); err != nil {
			continue
		}

		jobsClaimed = append(jobsClaimed, j)
	}

	jobsClaimedRows.Close()

	if err := tx.Commit(txCtx); err != nil {
		http.Error(w, "Error while commiting", http.StatusInternalServerError)
		return
	}

	for _, j := range jobsClaimed {
		jobs = append(jobs, j)
		_, _ = database.DB.Exec(ctx, `INSERT INTO jobs_history (job_id, event_type, worker_id, details) VALUES ($1, 'claimed', $2, $3)`, j.Id, workerId, json.RawMessage(`{}`))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NextJobResponse{Jobs: jobs})

	metrics.JobsInProgress.Inc()
}
