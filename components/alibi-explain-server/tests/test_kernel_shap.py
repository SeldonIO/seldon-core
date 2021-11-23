import json

import numpy as np
from sklearn.datasets import load_wine

from alibiexplainer.kernel_shap import KernelShap

from .make_test_models import make_kernel_shap


def test_kernel_shap():
    np.random.seed(0)

    alibi_model = make_kernel_shap()
    kernel_shap = KernelShap(alibi_model)
    wine = load_wine()
    X_test = wine.data
    explanation = kernel_shap.explain(X_test[0:1].tolist())
    exp_json = json.loads(explanation.to_json())
    assert exp_json["meta"]["name"] == "KernelShap"
