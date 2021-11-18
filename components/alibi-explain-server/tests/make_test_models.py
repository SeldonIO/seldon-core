import argparse
import os
from pathlib import Path
from typing import Optional

import numpy as np
import tensorflow as tf
from alibi.explainers import AnchorImage, KernelShap
from sklearn.datasets import load_wine
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
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


if __name__ == "__main__":
    _main()
