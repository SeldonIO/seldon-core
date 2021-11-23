# Copyright 2019 kubeflow.org.
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
# Originally copied from https://github.com/kubeflow/kfserving/blob/master/python/
# alibiexplainer/alibiexplainer/explainer.py
# and since modified
#

import json
import logging
import os
from typing import Any, Callable, Dict, Mapping

from alibi.api.interfaces import Explanation
from tensorflow import keras

import alibiexplainer.seldon_http as seldon
from alibiexplainer.ale import ALE
from alibiexplainer.anchor_images import AnchorImages
from alibiexplainer.anchor_tabular import AnchorTabular
from alibiexplainer.anchor_text import AnchorText
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.integrated_gradients import IntegratedGradients
from alibiexplainer.kernel_shap import KernelShap
from alibiexplainer.model import ExplainerModel
from alibiexplainer.tree_shap import TreeShap
from alibiexplainer.utils import ExplainerMethod, Protocol

SELDON_LOGLEVEL = os.environ.get("SELDON_LOGLEVEL", "INFO").upper()
logging.basicConfig(level=SELDON_LOGLEVEL)


class AlibiExplainer(ExplainerModel):
    def __init__(
        self,
        name: str,
        predict_fn: Callable,
        method: ExplainerMethod,
        config: Mapping,
        explainer: object = None,
        protocol: Protocol = Protocol.seldon_grpc,
        keras_model: keras.Model = None,
    ) -> None:
        super().__init__(name)
        self.method = method
        self.protocol = protocol
        logging.info("Protocol is %s", str(self.protocol))

        # Add type for first value to help pass mypy type checks
        if self.method is ExplainerMethod.anchor_tabular:
            self.wrapper: ExplainerWrapper = AnchorTabular(explainer, **config)
        elif self.method is ExplainerMethod.anchor_images:
            self.wrapper = AnchorImages(explainer, **config)
        elif self.method is ExplainerMethod.anchor_text:
            self.wrapper = AnchorText(predict_fn, explainer, **config)
        elif self.method is ExplainerMethod.kernel_shap:
            self.wrapper = KernelShap(explainer, **config)
        elif self.method is ExplainerMethod.integrated_gradients:
            self.wrapper = IntegratedGradients(keras_model, **config)
        elif self.method is ExplainerMethod.tree_shap:
            self.wrapper = TreeShap(explainer, **config)
        elif self.method is ExplainerMethod.ale:
            self.wrapper = ALE(explainer, **config)
        else:
            raise NotImplementedError

    def explain(self, request: Dict) -> Any:
        if (
            self.method is ExplainerMethod.anchor_tabular
            or self.method is ExplainerMethod.anchor_images
            or self.method is ExplainerMethod.anchor_text
            or self.method is ExplainerMethod.kernel_shap
            or self.method is ExplainerMethod.integrated_gradients
            or self.method is ExplainerMethod.tree_shap
            or self.method is ExplainerMethod.ale
        ):
            if self.protocol == Protocol.tensorflow_http:
                explanation: Explanation = self.wrapper.explain(request["instances"])
            else:
                rh = seldon.SeldonRequestHandler(request)
                response_list = rh.extract_request()
                explanation = self.wrapper.explain(response_list)
            explanation_as_json_str = explanation.to_json()
            logging.info("Explanation: %s", explanation_as_json_str)
            return json.loads(explanation_as_json_str)
        else:
            raise NotImplementedError
