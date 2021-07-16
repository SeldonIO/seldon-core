from alibiexplainer.ale import ALE
import kfserving
import os
import dill
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
import numpy as np
import json
from .utils import SKLearnServer
ALE_EXPLAINER_URI = "gs://seldon-models/sklearn/iris-0.23.2/ale_py37"
IRIS_MODEL_URI = "gs://seldon-models/sklearn/iris-0.23.2/lr_model"
EXPLAINER_FILENAME = "explainer.dill"


def test_ale():
    os.environ.clear()
    alibi_model = os.path.join(
        kfserving.Storage.download(ALE_EXPLAINER_URI), EXPLAINER_FILENAME
    )
    with open(alibi_model, "rb") as f:
        skmodel = SKLearnServer(IRIS_MODEL_URI)
        skmodel.load()
        alibi_model = dill.load(f)
        ale = ALE(skmodel.predict, alibi_model)
        data = load_iris()
        X = data.data
        y = data.target
        X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.25, random_state=42)
        np.random.seed(0)
        explanation = ale.explain(X_test.tolist())
        exp_json = json.loads(explanation.to_json())
        print(exp_json)