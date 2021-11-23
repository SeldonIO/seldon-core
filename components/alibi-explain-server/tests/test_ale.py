import json

import numpy as np
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split

from alibiexplainer.ale import ALE

from .make_test_models import make_ale
from .utils import SKLearnServer

IRIS_MODEL_URI = "gs://seldon-models/sklearn/iris-0.23.2/lr_model/*"


def test_ale():
    np.random.seed(0)

    alibi_model = make_ale()

    skmodel = SKLearnServer(IRIS_MODEL_URI)
    skmodel.load()
    data = load_iris()
    X = data.data
    y = data.target
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.25, random_state=42
    )
    ale = ALE(alibi_model)
    explanation = ale.explain(X_test.tolist())
    exp_json = json.loads(explanation.to_json())
    assert exp_json["meta"]["name"] == "ALE"
