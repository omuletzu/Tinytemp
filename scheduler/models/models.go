package models

import (
	"encoding/json"
	"time"
)

type Job struct {
	Id             int64           `json:"id"`
	JobType        string          `json:"job_type"`
	Payload        json.RawMessage `json:payload"`
	Status         string          `json:"status"`
	Attempts       int             `json:"attempts"`
	MaxAttempts    int             `json:"max_attempts"`
	IdempotencyKey *string         `json:"idempotency_key"`
	LockedUntil    *time.Time      `json:"locked_until"`
	NextRunAt      time.Time       `json:"next_run_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
