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


def test_cifar10_images():  # pylint: disable-msg=too-many-locals
    url = "https://storage.googleapis.com/seldon-models/alibi-detect/classifier/"
    path_model = os.path.join(url, "cifar10", "resnet32", "model.h5")
    save_path = tf.keras.utils.get_file("resnet32", path_model)
    model = tf.keras.models.load_model(save_path)

    # we drop the first batch dimension because AnchorImage expects a single image
    image_shape = model.get_layer(index=0).input_shape[0][1:]
    alibi_model = AnchorImage(predictor=model, image_shape=image_shape)
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
