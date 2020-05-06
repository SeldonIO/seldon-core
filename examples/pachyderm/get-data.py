from sklearn import datasets
import pandas as pd
import numpy as np


def main():
    print("Getting Iris Dataset")
    iris = datasets.load_iris()
    X, y = iris.data, iris.target

    data = pd.DataFrame(
        data=np.c_[iris["data"], iris["target"]],
        columns=iris["feature_names"] + ["target"],
    )

    data.to_csv("data.csv", index=False)
    print("Iris dataset saved to 'data.csv' file")


if __name__ == "__main__":
    main()
