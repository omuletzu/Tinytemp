CREATE TABLE IF NOT EXISTS jobs (
	id BIGSERIAL PRIMARY KEY,
	job_type VARCHAR(255),
	payload JSONB,
	status VARCHAR(255) DEFAULT 'pending',
	attempts INT DEFAULT 0,
	max_attempts INT DEFAULT 5,
	idempotency_key VARCHAR(255) UNIQUE,
	locked_until TIMESTAMP DEFAULT now(),
	next_run_at TIMESTAMP DEFAULT now(),
	created_at TIMESTAMP DEFAULT now(),
	updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS jobs_history (
	id BIGSERIAL PRIMARY KEY,
	job_id BIGINT REFERENCES jobs(id) ON DELETE CASCADE,
	event_type VARCHAR(255),
	worker_id VARCHAR(255),
	timestamp TIMESTAMP NOT NULL DEFAULT now(),
	details JSONB
);

CREATE INDEX IF NOT EXISTS jobs_idempotency_key_idx ON jobs(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS jobs_status_locked_until_idx ON jobs(status, locked_until);
CREATE INDEX IF NOT EXISTS jobs_status_locked_next_run_at_idx ON jobs(status, next_run_at);