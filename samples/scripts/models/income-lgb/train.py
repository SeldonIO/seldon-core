import numpy as np
import xgboost as xgb
from alibi.datasets import fetch_adult
from scipy.special import expit
import lightgbm as lgb
from pandas import DataFrame
import matplotlib.pyplot as plt

def wrap(arr):
    return np.ascontiguousarray(arr)


def main():
    adult = fetch_adult()
    adult.keys()

    data = adult.data
    target = adult.target
    target_names = adult.target_names
    feature_names = adult.feature_names
    category_map = adult.category_map

    np.random.seed(0)
    data_perm = np.random.permutation(np.c_[data, target])
    data = data_perm[:, :-1]
    target = data_perm[:, -1]

    idx = 30000
    X_train, y_train = data[:idx, :], target[:idx]
    X_test, y_test = data[idx + 1:, :], target[idx + 1:]

    d_train = lgb.Dataset(X_train, label=y_train)
    d_test = lgb.Dataset(X_test, label=y_test)

    params = {
        "max_bin": 512,
        "learning_rate": 0.05,
        "boosting_type": "gbdt",
        "objective": "binary",
        "metric": "binary_logloss",
        "num_leaves": 10,
        "verbose": -1,
        "min_data": 100,
        "boost_from_average": True
    }

    model = lgb.train(params, d_train, 10000, valid_sets=[d_test], early_stopping_rounds=50, verbose_eval=1000)
    model.save_model("model.bst")


if __name__ == "__main__":
    print("Building income model...")
    main()
