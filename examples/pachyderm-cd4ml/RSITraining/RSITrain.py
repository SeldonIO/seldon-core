import argparse

import joblib
import pandas as pd
from statsmodels.tsa.holtwinters import ExponentialSmoothing

parser = argparse.ArgumentParser(description='Train a model for iris classification.')
parser.add_argument('data_file_path', type=str, help='Path to csv file containing training set')
parser.add_argument('model_path', type=str, help='Path to output joblib file')
args = parser.parse_args()

print(f"Loading data set from {args.data_file_path}")
rsi = pd.read_csv(args.data_file_path, index_col="Date", parse_dates=True)

print("Training model...")
model = ExponentialSmoothing(rsi,trend='add').fit()
print("Model trained!")

print(f"Saving model in {args.model_path}")
joblib.dump(model, args.model_path)
print("Model saved!")