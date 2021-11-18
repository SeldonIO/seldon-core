import json
import os

import numpy as np
from alibi.datasets import fetch_adult

from alibiexplainer.tree_shap import TreeShap

from .make_test_models import make_tree_shap


def test_tree_shap():
    np.random.seed(0)

    alibi_model = make_tree_shap()
    tree_shap = TreeShap(alibi_model)
    adult = fetch_adult()
    X_test = adult.data[30001:, :]
    explanation = tree_shap.explain(X_test[0:1].tolist())
    exp_json = json.loads(explanation.to_json())
    assert exp_json["meta"]["name"] == "TreeShap"
