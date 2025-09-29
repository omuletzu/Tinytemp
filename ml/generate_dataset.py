import pandas as pd
import numpy as np

np.random.seed(42)

n_generated_samples = 10000

job_types = ["training", "inference", "evaluation", "preprocessing", "feature_extraction"]

heavier_models = [
    "resnet50", "resnet101", "vgg16", "vgg19", "efficientnet", "vit-base", "bert-base", 
    "bert-large", "gpt2", "gpt2-medium", "lstm", "gru", "yolo"
]

medium_models = [
    "mobilenet", "xgboost", "lightgbm", "catboost", "random_forest"
]

simple_models = [
    "linear_regression", "polynomial_regression", "logistic_regression",
    "support_vector_machine (svm)", "k-nearest_neighbors (k-nn)", 
    "naive_bayes", "k-means", "principal_component_analysis (pca)"
]


generated_data = []

for _ in range(n_generated_samples):
    job_type = np.random.choice(job_types)
    model = np.random.choice(heavier_models + medium_models + simple_models)

    dataset_size = int(np.random.lognormal(mean=10, sigma=1))

    batch_size = np.random.choice([16, 32, 64, 128], p=[0.4, 0.3, 0.2, 0.1])

    epochs = np.random.choice(range(1, 21), p = [0.3, 0.25, 0.15, 0.1, 0.2] + [0] * (15))

    worker_gpu = np.random.choice([4, 3, 2, 1], p=[0.1, 0.2, 0.3, 0.4])
    worker_cpu = np.random.choice([4, 8, 16, 32], p=[0.1, 0.4, 0.4, 0.1])
    worker_mem = np.random.choice([8192, 16384, 32768, 65536], p=[0.1, 0.3, 0.4, 0.2])

    if job_type == "training":
        runtime = epochs * (dataset_size / batch_size) * 2 * (32 / batch_size) / 100
    elif job_type == "inference":
        runtime = (dataset_size / batch_size) * (32 / batch_size) / 50
    elif job_type == "evaluation":
        runtime = (dataset_size / batch_size) * (32 / batch_size) / 40
    elif job_type == "preprocessing":
        runtime = dataset_size / 2500
    elif job_type == "feature_extraction":
        runtime = (dataset_size / batch_size) * (32 / batch_size) / 60

    if model in heavier_models:
        runtime *= 1.5
    elif model in medium_models:
        runtime *= 1.2

    gpu_multiplier = {
        1 : 1.5,
        2 : 1.2,
        3 : 1.0,
        4 : 0.8
    }

    runtime *= gpu_multiplier[worker_gpu]

    cpu_multiplier = {
        4 : 1.2,
        8 : 1.0,
        16 : 0.8,
        32 : 0.7
    }

    runtime *= cpu_multiplier[worker_cpu]

    mem_multiplier = {
        8192:1.1, 
        16384:1.0, 
        32768:0.9, 
        65536:0.85
    }
    
    runtime *= mem_multiplier[worker_mem]

    runtime *= np.random.uniform(0.9, 1.1)

    generated_data.append([
        job_type, model, dataset_size, batch_size, epochs, worker_cpu, worker_gpu, worker_mem, runtime
    ])

df = pd.DataFrame(generated_data, columns=[
    "job_type", "model", "dataset_size", "batch_size", "epochs", "worker_cpu", "worker_gpu", "worker_mem", "runtime"
])

df.to_csv("dataset.csv", index=False)