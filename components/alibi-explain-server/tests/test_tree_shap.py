from alibiexplainer.tree_shap import TreeShap
import kfserving
import os
import dill
from alibi.datasets import fetch_adult
import numpy as np
import json
import shap
ADULT_EXPLAINER_URI = "gs://seldon-models/xgboost/adult/tree_shap_py37_alibi_0.6.0"
EXPLAINER_FILENAME = "explainer.dill"


def test_tree_shap():
    os.environ.clear()
    alibi_model = os.path.join(
        kfserving.Storage.download(ADULT_EXPLAINER_URI), EXPLAINER_FILENAME
    )
    with open(alibi_model, "rb") as f:
        alibi_model = dill.load(f)
        tree_shap = TreeShap(alibi_model)
        adult = fetch_adult()
        X_test = adult.data[30001:, :]
        np.random.seed(0)
        explanation = tree_shap.explain(X_test[0:1].tolist())
        exp_json = json.loads(explanation.to_json())
        print(exp_json)