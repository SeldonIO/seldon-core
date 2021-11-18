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
# tests/test_anchor_images.py
# and modified since
#

import json
import os

import numpy as np
import tensorflow as tf
from alibi.explainers import AnchorImage

from alibiexplainer.anchor_images import AnchorImages

from .make_test_models import make_anchor_image


def test_cifar10_images():  # pylint: disable-msg=too-many-locals
    alibi_model = make_anchor_image()
    anchor_images = AnchorImages(alibi_model)

    _, test = tf.keras.datasets.cifar10.load_data()
    X_test, _ = test
    X_test = X_test.astype("float32") / 255
    idx = 12
    test_example = X_test[idx : idx + 1]

    np.random.seed(0)
    explanation = anchor_images.explain(test_example)
    exp_json = json.loads(explanation.to_json())
    assert exp_json["data"]["precision"] > 0.9
