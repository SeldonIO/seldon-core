import numpy as np
from sklearn.datasets import load_iris
from alibi.explainers import AnchorTabular

import requests


dataset = load_iris()
feature_names = dataset.feature_names
iris_data = dataset.data

model_url = "http://localhost:8003/seldon/seldon/iris/api/v1.0/predictions"


def predict_fn(X):
    data = {"data": {"ndarray": X.tolist()}}
    r = requests.post(model_url, json={"data": {"ndarray": [[1, 2, 3, 4]]}})
    return np.array(r.json()["data"]["ndarray"])

explainer = AnchorTabular(predict_fn, feature_names)
explainer.fit(iris_data, disc_perc=(25, 50, 75))

explainer.save("./explainer/")
