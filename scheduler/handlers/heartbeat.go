package handlers

import (
	"net/http"
	"strconv"
	"tinytemp/database"

	"github.com/go-chi/chi/v5"
)

func HeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	jobIdStr := chi.URLParam(req, "jobId")
	jobId, _ := strconv.ParseInt(jobIdStr, 10, 64)

	_, err := database.DB.Exec(ctx, `UPDATE jobs SET locked_until = now() + interval '30 seconds', updated_at = now() WHERE id = $1 AND status = 'in_progress'`, jobId)

	if err != nil {
		http.Error(w, "Cannot update job to be still in progress", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
