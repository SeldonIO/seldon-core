import argparse
import os

import joblib
import pandas as pd
from sklearn.pipeline import Pipeline
from sklearn.linear_model import LogisticRegression

# command line arguments
parser = argparse.ArgumentParser(description='Train a model for iris classification.')
parser.add_argument('indir', type=str, help='Input directory containing the training set')
parser.add_argument('outdir', type=str, help='Output directory for the trained model')
args = parser.parse_args()

# training set column names
cols = [
    "Sepal_Length",
    "Sepal_Width",
    "Petal_Length",
    "Petal_Width",
    "Species"
]

features = [
    "Sepal_Length",
    "Sepal_Width",
    "Petal_Length",
    "Petal_Width"
]

print(f"Loading iris data set from {args.indir}")
irisDF = pd.read_csv(os.path.join(args.indir, "iris.csv"), names=cols)

clf = LogisticRegression(solver="liblinear", multi_class="ovr")
p = Pipeline([("clf", clf)])
print("Training model...")
p.fit(irisDF[features], irisDF["Species"])
print("Model trained!")

print(f"Saving model in {args.outdir}")
joblib.dump(p, os.path.join(args.outdir, 'model.joblib'))
print("Model saved!")