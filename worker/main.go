package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type JobResp struct {
	Jobs []struct {
		Id      int64           `json:"id"`
		JobType string          `json:"job_type"`
		Payload json.RawMessage `json:"payload"`
	} `json:"jobs"`
}

type WorkerHardware struct {
	WorkerCPU int
	WorkerGPU int
	WorkerMEM int
}

func main() {
	godotenv.Load()
	rand.Seed(time.Now().UnixNano())

	api := os.Getenv("API_URL")
	workerId := os.Getenv("WORKER_ID")

	client := &http.Client{Timeout: 10 * time.Second}

	var workerHardware WorkerHardware = WorkerHardware{
		WorkerCPU: 8,
		WorkerGPU: 2,
		WorkerMEM: 8192,
	}

	bodyBytes, err := json.Marshal(workerHardware)
	if err != nil {
		panic(err)
	}

	baseDelay := 200 * time.Millisecond
	maxDelay := 5 * time.Second
	delay := baseDelay

	for {
		req, err := http.NewRequest("POST", api+"/next-job?worker_id="+workerId+"&capacity=1", bytes.NewBuffer(bodyBytes))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application.json")

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(delayWithJitter(delay))
			delay = minDuration(maxDelay, 2*delay)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Cannot read response body")
			time.Sleep(delayWithJitter(delay))
			continue
		}

		resp.Body.Close()

		delay = baseDelay

		if resp.StatusCode == http.StatusNoContent || len(bodyBytes) == 0 || resp.StatusCode != http.StatusOK {
			time.Sleep(delayWithJitter(delay))
			continue
		}

		var x JobResp
		if err := json.Unmarshal(body, &x); err != nil {
			log.Println("Cannot unmarshal response body")
			time.Sleep(delayWithJitter(delay))
			continue
		}

		if len(x.Jobs) == 0 {
			time.Sleep(delayWithJitter(delay))
			continue
		}

		job := x.Jobs[0]
		log.Println("Worker", workerId, "fetched", job.Id)

		sleepTime := rand.Intn(5000)
		sleepTimeDuration := time.Duration(sleepTime) * time.Millisecond

		time.Sleep(sleepTimeDuration)
		firstDigitTime := int(strconv.Itoa(sleepTime)[0] - '0')

		randFailJob := (rand.Intn(4) == 1)

		if randFailJob {

			failJob := map[string]any{"worker_id": workerId, "error": "simulated error", "duration_mils": firstDigitTime}
			data, _ := json.Marshal(failJob)

			_, err = http.Post(api+"/fail/"+strconv.FormatInt(job.Id, 10), "application/json", bytes.NewReader(data))

			if err == nil {
				log.Println("Worker", workerId, "failed job", job.Id)
			} else {
				log.Println("Worker", workerId, "couldn't fail job", job.Id)
			}

		} else {
			ackJob := map[string]any{"worker_id": workerId, "duration_mils": firstDigitTime}
			data, _ := json.Marshal(ackJob)

			_, err = http.Post(api+"/ack/"+strconv.FormatInt(job.Id, 10), "application/json", bytes.NewBuffer(data))

			if err == nil {
				log.Println("Worker", workerId, "completed job", job.Id)
			} else {
				log.Println("Worker", workerId, "couldn't complete job", job.Id)
			}
		}

		time.Sleep(delayWithJitter(delay))
	}
}

func delayWithJitter(x time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(x/2 + 1)))
	return x + jitter - (x / 4)
}

func minDuration(a time.Duration, b time.Duration) time.Duration {
	if a < b {
		return a
	}

	return b
}
