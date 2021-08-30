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
# Originally copied from https://github.com/kubeflow/kfserving/blob/master/python/alibiexplainer/alibiexplainer/explainer.py
# and since modifed
#

import json
import logging
from enum import Enum
from typing import List, Any, Mapping, Union, Dict
import numpy as np
from alibiexplainer.anchor_images import AnchorImages
from alibiexplainer.anchor_tabular import AnchorTabular
from alibiexplainer.anchor_text import AnchorText
from alibiexplainer.kernel_shap import KernelShap
from alibiexplainer.tree_shap import TreeShap
from alibiexplainer.integrated_gradients import IntegratedGradients
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.proto import prediction_pb2
from alibiexplainer.proto import prediction_pb2_grpc
from alibi.api.interfaces import Explanation
import grpc
import tensorflow as tf
import alibiexplainer.seldon_http as seldon
import alibiexplainer.v2_http as v2
import requests
import os
from alibiexplainer.model import ExplainerModel
from tensorflow import keras

SELDON_LOGLEVEL = os.environ.get('SELDON_LOGLEVEL', 'INFO').upper()
logging.basicConfig(level=SELDON_LOGLEVEL)
GRPC_MAX_MSG_LEN = 1000000000
TENSORFLOW_PREDICTOR_URL_FORMAT = "http://{0}/v1/models/{1}:predict"
SELDON_PREDICTOR_URL_FORMAT = "http://{0}/api/v0.1/predictions"
V2_PREDICTOR_URL_FORMAT = "http://{0}/v2/models/{1}/infer"

class Protocol(Enum):
    tensorflow_http = "tensorflow.http"
    seldon_http = "seldon.http"
    seldon_grpc = "seldon.grpc"
    v2_http = "kfserving.http"

    def __str__(self):
        return self.value

class ExplainerMethod(Enum):
    anchor_tabular = "AnchorTabular"
    anchor_images = "AnchorImages"
    anchor_text = "AnchorText"
    kernel_shap = "KernelShap"
    integrated_gradients = "IntegratedGradients"
    tree_shap = "TreeShap"

    def __str__(self):
        return self.value


class AlibiExplainer(ExplainerModel):
    def __init__(self,
                 name: str,
                 predictor_host: str,
                 method: ExplainerMethod,
                 config: Mapping,
                 explainer: object = None,
                 protocol: Protocol = Protocol.seldon_grpc,
                 tf_data_type: str = None,
                 keras_model: keras.Model = None ):
        super().__init__(name)
        self.predictor_host = predictor_host
        logging.info("Predict URL set to %s", self.predictor_host)
        self.method = method
        self.protocol = protocol
        self.tf_data_type = tf_data_type
        logging.info("Protocol is %s",str(self.protocol))
        self.v2_name = None
        self.v2_type = None
        self.v2_model_name = None

        # Add type for first value to help pass mypy type checks
        if self.method is ExplainerMethod.anchor_tabular:
            self.wrapper:ExplainerWrapper = AnchorTabular(self._predict_fn, explainer, **config)
        elif self.method is ExplainerMethod.anchor_images:
            self.wrapper = AnchorImages(self._predict_fn, explainer, **config)
        elif self.method is ExplainerMethod.anchor_text:
            self.wrapper = AnchorText(self._predict_fn, explainer, **config)
        elif self.method is ExplainerMethod.kernel_shap:
            self.wrapper = KernelShap(self._predict_fn, explainer, **config)
        elif self.method is ExplainerMethod.integrated_gradients:
            self.wrapper = IntegratedGradients(keras_model, **config)
        elif self.method is ExplainerMethod.tree_shap:
            self.wrapper = TreeShap(explainer, **config)
        else:
            raise NotImplementedError

    def load(self):
        pass

    def _predict_fn(self, arr: Union[np.ndarray, List]) -> np.ndarray:
        print(arr)
        if type(arr) == list:
            arr = np.array(arr)
        if self.protocol == Protocol.seldon_grpc:
            return self._grpc(arr)
        elif self.protocol == Protocol.seldon_http:
            payload = seldon.create_request(arr, seldon.SeldonPayload.NDARRAY)
            response_raw = requests.post(SELDON_PREDICTOR_URL_FORMAT.format(self.predictor_host), json=payload)
            if response_raw.status_code == 200:
                rh = seldon.SeldonRequestHandler(response_raw.json())
                response_list = rh.extract_request()
                return np.array(response_list)
            else:
                raise Exception(
                    "Failed to get response from model return_code:%d" % response_raw.status_code)
        elif self.protocol == Protocol.tensorflow_http:
            instances = []
            for req_data in arr:
                if isinstance(req_data, np.ndarray):
                    instances.append(req_data.tolist())
                else:
                    instances.append(req_data)
            request = {"instances": instances}
            print(self.predictor_host)
            response = requests.post(
                TENSORFLOW_PREDICTOR_URL_FORMAT.format(self.predictor_host, self.v2_model_name),
                json.dumps(request)
            )
            if response.status_code != 200:
                raise Exception(
                    "Failed to get response from model return_code:%d" % response.status_code)
            return np.array(response.json()["predictions"])
        elif self.protocol == Protocol.v2_http:
            rh = v2.KFServingV2RequestHandler()
            request = rh.create_request(arr, self.v2_name, self.v2_type)
            logging.info("url %s",V2_PREDICTOR_URL_FORMAT.format(self.predictor_host, self.v2_model_name))
            response = requests.post(
                V2_PREDICTOR_URL_FORMAT.format(self.predictor_host, self.v2_model_name),
                json.dumps(request)
            )
            if response.status_code != 200:
                raise Exception(
                    "Failed to get response from model return_code:%d" % response.status_code)
            response_json = response.json()
            arr = rh.extract_response(response_json)
            return arr


    def explain(self, request: Dict, model_name=None) -> Any:
        if self.method is ExplainerMethod.anchor_tabular or self.method is ExplainerMethod.anchor_images or \
                self.method is ExplainerMethod.anchor_text or self.method is ExplainerMethod.kernel_shap or \
                self.method is ExplainerMethod.integrated_gradients or self.method is ExplainerMethod.tree_shap:
            if self.protocol == Protocol.tensorflow_http:
                self.v2_model_name = model_name
                explanation: Explanation = self.wrapper.explain(request["instances"])
            elif self.protocol == Protocol.v2_http:
                logging.info("model name %s", model_name)
                rh = v2.KFServingV2RequestHandler()
                self.v2_model_name = model_name
                self.v2_name = rh.extract_name(request)
                self.v2_type = rh.extract_type(request)
                logging.info("v2 name from inputs %s:", self.v2_name)
                response_list = rh.extract_request(request)
                explanation = self.wrapper.explain(response_list)
            else:
                rh = seldon.SeldonRequestHandler(request)
                response_list = rh.extract_request()
                explanation = self.wrapper.explain(response_list)
            explanationAsJsonStr = explanation.to_json()
            logging.info("Explanation: %s", explanationAsJsonStr)
            return json.loads(explanationAsJsonStr)
        else:
            raise NotImplementedError

    def _grpc(self, arr: np.array) -> np.array:
        options = [
            ('grpc.max_send_message_length', GRPC_MAX_MSG_LEN),
            ('grpc.max_receive_message_length', GRPC_MAX_MSG_LEN)]
        channel = grpc.insecure_channel(self.predictor_host, options)
        stub = prediction_pb2_grpc.SeldonStub(channel)
        if self.tf_data_type is not None:
            datadef = prediction_pb2.DefaultData(
                tftensor=tf.make_tensor_proto(arr, self.tf_data_type))
        else:
            datadef = prediction_pb2.DefaultData(
                tftensor=tf.make_tensor_proto(arr))
        request = prediction_pb2.SeldonMessage(data=datadef)
        response = stub.Predict(request=request)
        arr_resp = tf.make_ndarray(response.data.tftensor)
        return arr_resp