# Tinytemp

**ML Job Scheduler**

[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev/)  
[![Python](https://img.shields.io/badge/Python-3.10-yellow)](https://www.python.org/)  
[![Docker](https://img.shields.io/badge/Docker-✓-2496ED)](https://www.docker.com/)  
[![Kubernetes](https://img.shields.io/badge/Kubernetes-✓-326CE5)](https://kubernetes.io/)  
[![Postgres](https://img.shields.io/badge/PostgreSQL-✓-336791)](https://www.postgresql.org/)  
[![Prometheus](https://img.shields.io/badge/Monitoring-Prometheus-orange)](https://prometheus.io/)  

**Tinytemp** is a distributed job scheduler optimized for Machine Learning tasks, designed to efficiently assign jobs to workers based on a **predicted runtime**, including **worker 
hardware capabilities** and **task specification**.

It includes features like:
- Task prioritization
- Retries with backoff
- Monitoring and full observability via Prometheus metrics.

## Table of Contents

- [Project Overview](#project-overview)
- [Tech Stack](#tech-stack)
- [Components](#components)
  - [Scheduler](#scheduler)
  - [Worker](#worker)
  - [Predictor (ML Model)](#predictor-ml-model)
  - [Database](#database)
- [Machine Learning](#machine-learning)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Screenshots](#screenshots)

---

## Project Overview

Tinytemp allows enqueueing ML tasks such as training, evaluation, inference and more. Workers pull jobs from the scheduler, execute them, and report results. Tasks are scheduled intelligently based on:

- Worker hardware (CPU, GPU, memory)
- Job parameters (dataset size, model type, batch size, epochs)
- Predicted runtime from an ML model

The system ensures:

- No two workers execute the same job simultaneously
- Automatic retries with exponential backoff
- Task prioritization to minimize total runtime
- Real-time metrics via Prometheus

## Tech Stack

- **Backend:** Go (Scheduler & Worker)
- **Machine Learning:** Python, scikit-learn, XGBoost
- **Database:** PostgreSQL
- **API:** FastAPI for ML predictor
- **Containerization:** Docker
- **Orchestration:** Kubernetes (Minikube)
- **Monitoring:** Prometheus

## Components

### Scheduler

- Written in **Go**, exposes REST API endpoints:
  - `/enqueue` – enqueue new jobs
  - `/next-job` – assign next jobs to workers based on priority
  - `/ack/{jobId}` – mark job as succeeded
  - `/fail/{jobId}` – mark job as failed and handle retries
  - `/heartbeat/{jobId}` – keep job in progress alive
- Integrates with **PostgreSQL** for job persistence
- Implements **task prioritization** using predicted runtime and task age
- Tracks job events in `jobs_history` table
- Containerized via Docker and deployable in Kubernetes

### Worker

- Written in **Go**
- Polls scheduler via `/next-job`
- Simulates job execution
- Reports results via `/ack` or `/fail`
- Implements exponential backoff with jitter for retries
- Multiple replicas can run concurrently

### Predictor (ML Model)

- **Python + FastAPI**
- Loads a trained Random Forest pipeline to predict job runtime
- Endpoint: `/predict`
- Trained on a synthetic dataset simulating different ML tasks
- Hyperparameter optimized with RandomizedSearchCV
- Supports multiple input features including `job_type`, `model`, `dataset_size`, `epochs`, `batch_size`, and worker hardware

### Database

- **PostgreSQL**
- Tables:
  - `jobs`: stores enqueued tasks with status, attempts, payload, etc.
  - `jobs_history`: stores task events like enqueued, claimed, acked, failed
- Indexed for fast queries and to avoid task duplication

## Machine Learning

The ML model predicts runtime for a task, enabling the scheduler to assign jobs efficiently. Key details:

- **Model:** Random Forest Regressor (best among tested models: XGBoost, Polynomial Regression)
- **Hyperparameter Tuning:** RandomizedSearchCV
- **Input Features:**
  - Job type (training, evaluation, inference)
  - Model name (ResNet50, etc.)
  - Dataset size, batch size, epochs
  - Worker CPU, GPU, memory
- **Output:** Estimated runtime in seconds
- Stored as `runtime_predictor.pkl` and served via FastAPI

### Deployment

All components of Tinytemp (scheduler, workers, predictor, and database) were containerized and built in the local Docker registry. Kubernetes manifests were applied to deploy the system locally using Minikube, orchestrating all services and enabling full end-to-end execution of ML tasks.

Example steps:

```bash
# Build images
docker build -t scheduler ./scheduler
docker build -t worker ./worker
docker build -t predictor ./predictor

# Apply Kubernetes manifests
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/predictor.yaml
kubectl apply -f k8s/scheduler.yaml
kubectl apply -f k8s/worker.yaml
```

### Monitoring

Prometheus metrics exposed at `/metrics`
Tracks:
- Total jobs enqueued
- Jobs in progress
- Jobs succeeded
- Jobs failed
- Jobs dead-lettered (DLQ)
- Job processing duration

### Screenshots

**1. Job Enqueuing**  
Example of a job being successfully enqueued via the Tinytemp API.

<img width="1396" height="740" alt="Job Enqueuing" src="https://github.com/user-attachments/assets/38db84e3-49f5-4a12-b462-0af56c918d8b" />

**2. All Available Jobs**  
Current jobs in the database, showing pending and in-progress tasks.  

<img width="1902" height="161" alt="Jobs History" src="https://github.com/user-attachments/assets/38346e43-e37b-49d5-81bc-205dc413f82d" />
<img width="1898" height="153" alt="All Jobs 2" src="https://github.com/user-attachments/assets/979a63b3-d037-457f-869f-f263dfbb5733" />

**3. Jobs History**  
Demonstrates the job lifecycle tracking, including succeeded, failed, and retried jobs, stored in PostgreSQL.

<img width="1303" height="163" alt="All Jobs 1" src="https://github.com/user-attachments/assets/ad18751b-b0c9-48c7-b331-a929add057ad" />  

