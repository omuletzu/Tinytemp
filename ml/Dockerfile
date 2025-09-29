FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt /app/requirements.txt

RUN pip install --no-cache-dir --upgrade -r /app/requirements.txt

COPY runtime_predictor.pkl /app/runtime_predictor.pkl

COPY runtime_predictor.py /app/runtime_predictor.py

EXPOSE 8001

CMD ["uvicorn", "runtime_predictor:app", "--host", "0.0.0.0", "--port", "8001"]