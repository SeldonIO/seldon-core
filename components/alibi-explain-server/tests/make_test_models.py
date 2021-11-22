import argparse
import os
from pathlib import Path
from typing import Optional

import numpy as np
import tensorflow as tf
import xgboost
from alibi.datasets import fetch_adult
from alibi.explainers import ALE, AnchorImage, AnchorTabular, KernelShap, TreeShap
from joblib import dump
from sklearn.compose import ColumnTransformer
from sklearn.datasets import load_iris, load_wine
from sklearn.ensemble import RandomForestClassifier
from sklearn.impute import SimpleImputer
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import train_test_split
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, StandardScaler
from sklearn.svm import SVC


def make_anchor_image(dirname: Optional[Path] = None) -> AnchorImage:
    url = "https://storage.googleapis.com/seldon-models/alibi-detect/classifier/"
    path_model = os.path.join(url, "cifar10", "resnet32", "model.h5")
    save_path = tf.keras.utils.get_file("resnet32", path_model)
    model = tf.keras.models.load_model(save_path)

    # we drop the first batch dimension because AnchorImage expects a single image
    image_shape = model.get_layer(index=0).input_shape[0][1:]
    alibi_model = AnchorImage(predictor=model, image_shape=image_shape)

    if dirname is not None:
        alibi_model.save(dirname)
    return alibi_model


def make_kernel_shap(dirname: Optional[Path] = None) -> KernelShap:
    np.random.seed(0)

    # load data
    wine = load_wine()
    data = wine.data
    target = wine.target
    target_names = wine.target_names
    feature_names = wine.feature_names

    # train classifier
    X_train, X_test, y_train, y_test = train_test_split(
        data, target, test_size=0.2, random_state=0
    )

    scaler = StandardScaler().fit(X_train)
    X_train_norm = scaler.transform(X_train)
    X_test_norm = scaler.transform(X_test)

    classifier = SVC(
        kernel="rbf",
        C=1,
        gamma=0.1,
        decision_function_shape="ovr",
        # n_cls trained with data from one class as positive
        # and remainder of data as neg
        random_state=0,
    )
    classifier.fit(X_train_norm, y_train)

    # build kernel shap model
    pred_fcn = classifier.decision_function
    svm_explainer = KernelShap(pred_fcn)
    svm_explainer.fit(X_train_norm)

    if dirname is not None:
        svm_explainer.save(dirname)
    return svm_explainer


def make_tree_shap(dirname: Optional[Path] = None) -> TreeShap:
    np.random.seed(0)

    # get X_train for explainer fit
    adult = fetch_adult()
    data = adult.data
    target = adult.target
    data_perm = np.random.permutation(np.c_[data, target])
    data = data_perm[:, :-1]
    target = data_perm[:, -1]
    idx = 30000
    X_train, y_train = data[:idx, :], target[:idx]
    X_test, y_test = data[idx + 1 :, :], target[idx + 1 :]

    d_train = xgboost.DMatrix(X_train, label=y_train)
    d_test = xgboost.DMatrix(X_test, label=y_test)

    params = {
        "eta": 0.01,
        "objective": "binary:logistic",
        "subsample": 0.5,
        "base_score": np.mean(y_train),
        "eval_metric": "logloss",
    }
    model = xgboost.train(
        params,
        d_train,
        5000,
        evals=[(d_test, "test")],
        verbose_eval=100,
        early_stopping_rounds=20,
    )

    tree_explainer = TreeShap(model, model_output="raw", task="classification")
    tree_explainer.fit(X_train)

    if dirname is not None:
        tree_explainer.save(dirname)
    return tree_explainer


def make_ale(dirname: Optional[Path] = None) -> ALE:
    data = load_iris()
    feature_names = data.feature_names
    target_names = data.target_names
    X = data.data
    y = data.target
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.25, random_state=42
    )

    # train model
    lr = LogisticRegression(max_iter=200)
    lr.fit(X_train, y_train)

    # create explainer
    logit_fun_lr = lr.decision_function
    logit_ale_lr = ALE(
        logit_fun_lr, feature_names=feature_names, target_names=target_names
    )

    if dirname is not None:
        logit_ale_lr.save(dirname)
    return logit_ale_lr


def make_anchor_tabular(dirname: Optional[Path] = None) -> AnchorTabular:
    # train model
    iris_data = load_iris()

    clf = LogisticRegression(solver="liblinear", multi_class="ovr")
    clf.fit(iris_data.data, iris_data.target)

    # create explainer
    explainer = AnchorTabular(clf.predict, feature_names=iris_data.feature_names)
    explainer.fit(iris_data.data, disc_perc=(25, 50, 75))

    if dirname is not None:
        explainer.save(dirname)
    return explainer


def make_anchor_tabular_income(dirname: Optional[Path] = None) -> AnchorTabular:
    # adapted from:
    # https://docs.seldon.io/projects/alibi/en/latest/examples/anchor_tabular_adult.html
    np.random.seed(0)

    # prepare data
    adult = fetch_adult()
    data = adult.data
    target = adult.target
    feature_names = adult.feature_names
    category_map = adult.category_map

    data_perm = np.random.permutation(np.c_[data, target])
    data = data_perm[:, :-1]
    target = data_perm[:, -1]

    # build model
    idx = 30000
    X_train, Y_train = data[:idx, :], target[:idx]
    X_test, Y_test = data[idx + 1 :, :], target[idx + 1 :]

    ordinal_features = [
        x for x in range(len(feature_names)) if x not in list(category_map.keys())
    ]
    ordinal_transformer = Pipeline(
        steps=[
            ("imputer", SimpleImputer(strategy="median")),
            ("scaler", StandardScaler()),
        ]
    )

    categorical_features = list(category_map.keys())
    categorical_transformer = Pipeline(
        steps=[
            ("imputer", SimpleImputer(strategy="median")),
            ("onehot", OneHotEncoder(handle_unknown="ignore")),
        ]
    )

    preprocessor = ColumnTransformer(
        transformers=[
            ("num", ordinal_transformer, ordinal_features),
            ("cat", categorical_transformer, categorical_features),
        ]
    )

    clf = RandomForestClassifier(n_estimators=50)

    model_pipeline = Pipeline(
        steps=[
            ("preprocess", preprocessor),
            ("classifier", clf),
        ]
    )

    model_pipeline.fit(X_train, Y_train)

    explainer = AnchorTabular(
        model_pipeline.predict, feature_names, categorical_names=category_map, seed=1
    )

    explainer.fit(X_train, disc_perc=[25, 50, 75])

    if dirname is not None:
        explainer.save(dirname)
    return explainer


def _main():
    args_parser = argparse.ArgumentParser(add_help=False)
    args_parser.add_argument(
        "--model",
        type=str,
        help="The model to create",
    )
    args_parser.add_argument(
        "--model_dir",
        type=Path,
        help="Where to save",
    )
    args = args_parser.parse_args()
    model_name = args.model
    model_dir = args.model_dir
    if model_name == "anchor_image":
        make_anchor_image(model_dir)
    elif model_name == "kernel_shap":
        make_kernel_shap(model_dir)
    elif model_name == "tree_shap":
        make_tree_shap(model_dir)
    elif model_name == "ale":
        make_ale(model_dir)
    elif model_name == "anchor_tabular":
        make_anchor_tabular(model_dir)
    elif model_name == "anchor_tabular_income":
        make_anchor_tabular_income(model_dir)


if __name__ == "__main__":
    _main()
