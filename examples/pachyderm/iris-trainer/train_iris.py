import os
import yaml
import joblib
import pandas as pd
from sklearn.pipeline import Pipeline
from sklearn.linear_model import LogisticRegression


# /pfs/iris-data and /pfs/outs are special directories mounted by Pachyderm
INPUT_FILE = os.path.join("/pfs/iris-data", "data.csv")
OUTPUT_FILE = os.path.join("/pfs/out", "model.joblib")


def main():
    clf = LogisticRegression(solver="liblinear", multi_class="ovr")
    p = Pipeline([("clf", clf)])
    print("Training model...")
    p.fit(X, y)
    print("Model trained!")

    print(f"Saving model in {OUTPUT_FILE}")
    joblib.dump(p, OUTPUT_FILE)
    print("Model saved!")


if __name__ == "__main__":
    print(f"Loading iris data set from {INPUT_FILE}")
    data = pd.read_csv(INPUT_FILE)
    y = data["target"]
    X = data[(x for x in data.columns if x != "target")]
    print("Dataset loaded!")

    main()
