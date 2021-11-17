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
# Copied from https://github.com/kubeflow/kfserving/blob/master/python/alibiexplainer/
# tests/test_anchor_tabular.py
# and modified since
#

import json
import tempfile

import numpy as np
from alibi.saving import load_explainer

from alibiexplainer.anchor_tabular import AnchorTabular

from .utils import SKLearnServer, download_from_gs

# to recreate these artifacts, use notebooks/explainer_examples_v2.ipynb
# TODO: move to a python runnable
IRIS_MODEL_URI = "gs://seldon-models/v1.11.0-dev/sklearn/iris/*"
# note: nothing special about v2
IRIS_EXPLAINER_URI = "gs://seldon-models/alibi/iris_anchor_tabular_explainer_v2/*"


def test_anchor_tabular():
    skmodel = SKLearnServer(IRIS_MODEL_URI)
    skmodel.load()

    with tempfile.TemporaryDirectory() as local_alibi_model:
        download_from_gs(IRIS_EXPLAINER_URI, local_alibi_model)
        alibi_model = load_explainer(predictor=skmodel.predict, path=local_alibi_model)
        anchor_tabular = AnchorTabular(alibi_model)

    test_data = np.array([[5.964, 4.006, 2.081, 1.031]])
    explanation = anchor_tabular.explain(test_data)
    explanation_json = json.loads(explanation.to_json())
    assert explanation_json["meta"]["name"] == "AnchorTabular"
