from alibiexplainer.kernel_shap import KernelShap
import kfserving
import os
import dill
from sklearn.datasets import load_wine
import numpy as np
import json
from .utils import SKLearnServer
WINE_EXPLAINER_URI = "gs://seldon-models/sklearn/wine/kernel_shap_py36_alibi_0.5.5"
WINE_MODEL_URI = "gs://seldon-models/sklearn/wine/model-py36-0.23.2"
EXPLAINER_FILENAME = "explainer.dill"


def test_kernel_shap():

    alibi_model = os.path.join(
        kfserving.Storage.download(WINE_EXPLAINER_URI), EXPLAINER_FILENAME
    )
    with open(alibi_model, "rb") as f:
        skmodel = SKLearnServer(WINE_MODEL_URI)
        skmodel.load()
        alibi_model = dill.load(f)
        kernel_shap = KernelShap(skmodel.predict, alibi_model)
        wine = load_wine()
        X_test = wine.data
        np.random.seed(0)
        explanation = kernel_shap.explain(X_test[0:1].tolist())
        exp_json = json.loads(explanation.to_json())
        print(exp_json)
