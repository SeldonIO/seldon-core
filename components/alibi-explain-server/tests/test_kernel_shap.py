from alibiexplainer.kernel_shap import KernelShap
import kfserving
import os
import dill
from alibi.datasets import fetch_adult
import numpy as np
import json
from .utils import SKLearnServer
ADULT_EXPLAINER_URI = "gs://seldon-models/sklearn/adult_shap/kernelshap/py36"
ADULT_MODEL_URI = "gs://seldon-models/sklearn/adult_shap/model"
EXPLAINER_FILENAME = "explainer.dill"


def test_kernel_shap():
    os.environ.clear()
    alibi_model = os.path.join(
        kfserving.Storage.download(ADULT_EXPLAINER_URI), EXPLAINER_FILENAME
    )
    with open(alibi_model, "rb") as f:
        skmodel = SKLearnServer(ADULT_MODEL_URI)
        skmodel.load()
        alibi_model = dill.load(f)
        kernel_shap = KernelShap(skmodel.predict, alibi_model)
        adult = fetch_adult()
        X_test = adult.data[30001:, :]
        np.random.seed(0)
        explanation = kernel_shap.explain(X_test[0:1].tolist())
        exp_json = json.loads(explanation.to_json())
        print(exp_json)
        #assert exp_json["data"]["anchor"][0] == "Age <= 28.00"
        #assert exp_json["data"]["anchor"][1] == "Marital Status = Never-Married"