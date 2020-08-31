# Copyright 2020 kubeflow.org.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Copied from https://github.com/kubeflow/kfserving/blob/master/python/alibiexplainer/tests/test_anchor_tabular.py
# and modified since
#

from alibiexplainer.anchor_tabular import AnchorTabular
import kfserving
import os
import dill
from alibi.datasets import fetch_adult
import numpy as np
import json
from .utils import SKLearnServer
ADULT_EXPLAINER_URI = "gs://seldon-models/sklearn/income/alibi/0.4.0"
ADULT_MODEL_URI = "gs://seldon-models/sklearn/income/model-0.23.2"
EXPLAINER_FILENAME = "explainer.dill"


def test_anchor_tabular():
    os.environ.clear()
    alibi_model = os.path.join(
        kfserving.Storage.download(ADULT_EXPLAINER_URI), EXPLAINER_FILENAME
    )
    with open(alibi_model, "rb") as f:
        skmodel = SKLearnServer(ADULT_MODEL_URI)
        skmodel.load()
        alibi_model = dill.load(f)
        anchor_tabular = AnchorTabular(skmodel.predict, alibi_model)
        adult = fetch_adult()
        X_test = adult.data[30001:, :]
        np.random.seed(0)
        explanation = anchor_tabular.explain(X_test[0:1].tolist())
        exp_json = json.loads(explanation.to_json())
        assert exp_json["data"]["anchor"][0] == "Marital Status = Never-Married"
