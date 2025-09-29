from fastapi import FastAPI
import joblib
import pandas as pd
import uvicorn
from pydantic import BaseModel

app = FastAPI()

model = joblib.load("runtime_predictor.pkl")

NUMERIC_COLS = ["dataset_size", "batch_size", "epochs", "worker_cpu", "worker_gpu", "worker_mem"]
CATEGORICAL_COLS = ["job_type", "model"]
ALL_COLS = NUMERIC_COLS + CATEGORICAL_COLS

class PredictRequest(BaseModel):
    dataset_size: int
    batch_size: int
    epochs: int
    worker_cpu: int = 8
    worker_gpu: int = 1
    worker_mem: int = 8192
    job_type: str
    model: str


@app.post("/predict")
def predict_job(req: PredictRequest):
    row = {col: getattr(req, col) for col in ALL_COLS}

    df = pd.DataFrame([row], columns=ALL_COLS)

    prediction = model.predict(df)
    return {
        "runtime": float(prediction[0])
    }

if __name__ == "__main__":
    uvicorn.run("exposed-model:app", host="127.0.0.1", port=8001, reload=False)