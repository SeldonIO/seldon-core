import matplotlib.pyplot as plt
import numpy as np
import xgboost as xgb

from alibi.datasets import fetch_adult
from copy import deepcopy
from functools import partial
from itertools import chain, product
from scipy.special import expit
invlogit=expit
from sklearn.metrics import accuracy_score, confusion_matrix
from tqdm import tqdm


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

    dtrain = xgb.DMatrix(
        wrap(X_train),
        label=wrap(y_train),
        feature_names=feature_names,
    )

    dtest = xgb.DMatrix(wrap(X_test), label=wrap(y_test), feature_names=feature_names)

    learning_params = {
        'objective': 'binary:logitraw',
        'seed': 42,
        'eval_metric': ['auc', 'logloss']  # metrics computed for specified dataset
    }

    params = {
        'scale_pos_weight': 2,
        'min_child_weight': 0.1,
        'max_depth': 3,
        'gamma': 0.01,
        'boost_rounds': 541,
    }

    params.update(learning_params)

    if 'boost_rounds' in params:
        boost_rounds = params.pop('boost_rounds')

    model = xgb.train(
        params,
        dtrain,
        num_boost_round=boost_rounds,
        evals=[(dtrain, "Train"), (dtest, "Test")],
    )

    model.save_model('model.bst')


if __name__ == "__main__":
    print("Building income model...")
    main()
