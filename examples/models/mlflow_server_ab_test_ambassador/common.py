import pandas as pd
import numpy as np

from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score


def eval_metrics(actual, pred):
    rmse = np.sqrt(mean_squared_error(actual, pred))
    mae = mean_absolute_error(actual, pred)
    r2 = r2_score(actual, pred)
    return rmse, mae, r2


def read_data():
    data = pd.read_csv("wine-quality.csv")
    data.head()

    # We normalize the inputs to both the SparkML & TensorFlow models
    # so that they have the same input schema.
    for col in data.columns[:-1]:
        data[col] = (data[col] - data[col].mean()) / data[col].std()

    return data
