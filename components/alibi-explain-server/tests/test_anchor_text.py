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
# Copied from https://github.com/kubeflow/kfserving/blob/master/python/alibiexplainer/tests/test_anchor_text.py
# and modified since
#

from alibiexplainer.anchor_text import AnchorText
import os
from alibi.datasets import fetch_movie_sentiment
from .utils import SKLearnServer
import json
import numpy as np

MOVIE_MODEL_URI = "gs://seldon-models/sklearn/moviesentiment_sklearn_0.24.2"


def test_anchor_text():
    os.environ.clear()
    skmodel = SKLearnServer(MOVIE_MODEL_URI)
    skmodel.load()
    movies = fetch_movie_sentiment()
    anchor_text = AnchorText(skmodel.predict, None)

    np.random.seed(0)
    explanation = anchor_text.explain(movies.data[4:5])
    exp_json = json.loads(explanation.to_json())
    print(exp_json["data"]["anchor"])
