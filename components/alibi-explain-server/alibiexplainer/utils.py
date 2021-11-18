import json
import logging
import os
from enum import Enum
from typing import Callable, List, Optional, Union

import grpc
import numpy as np
import requests
import tensorflow as tf
from alibi.api.interfaces import Explainer
from alibi.saving import load_explainer
from keras.models import Model
from tensorflow import keras

import alibiexplainer.seldon_http as seldon
from alibiexplainer.proto import prediction_pb2, prediction_pb2_grpc

SELDON_LOGLEVEL = os.environ.get("SELDON_LOGLEVEL", "INFO").upper()
logging.basicConfig(level=SELDON_LOGLEVEL)
GRPC_MAX_MSG_LEN = 1000000000

TENSORFLOW_PREDICTOR_URL_FORMAT = "http://{0}/v1/models/{1}:predict"
SELDON_PREDICTOR_URL_FORMAT = "http://{0}/api/v0.1/predictions"

_KERAS_MODEL_FILENAME = "model.h5"
_EXPLAINER_FILENAME = "explainer.dill"


class Protocol(Enum):
    tensorflow_http = "tensorflow.http"
    seldon_http = "seldon.http"
    seldon_grpc = "seldon.grpc"

    def __str__(self):
        return self.value


class ExplainerMethod(Enum):
    anchor_tabular = "AnchorTabular"
    anchor_images = "AnchorImages"
    anchor_text = "AnchorText"
    kernel_shap = "KernelShap"
    integrated_gradients = "IntegratedGradients"
    tree_shap = "TreeShap"
    ale = "ALE"

    def __str__(self):
        return self.value


def is_persisted_keras(dirname: str) -> bool:
    return os.path.exists(os.path.join(dirname, _KERAS_MODEL_FILENAME))


def get_persisted_keras(dirname: str) -> Model:
    keras_path = os.path.join(dirname, _KERAS_MODEL_FILENAME)
    with open(keras_path, "rb") as f:
        logging.info(f"Loading Keras model from {dirname}")
        return keras.models.load_model(keras_path)


def is_persisted_explainer(dirname: str) -> bool:
    return os.path.exists(os.path.join(dirname, _EXPLAINER_FILENAME))


def get_persisted_explainer(dirname, predict_fn: Callable) -> Explainer:
    logging.info(f"Loading Alibi model from {dirname}")
    return load_explainer(predictor=predict_fn, path=dirname)


def construct_predict_fn(
    predictor_host: str,
    model_name: str,
    protocol: Protocol = Protocol.seldon_grpc,
    tf_data_type: str = None,
) -> Callable:
    def _predict_fn(arr: Union[np.ndarray, List]) -> np.ndarray:
        if type(arr) == list:
            arr = np.array(arr)
        if protocol == Protocol.seldon_grpc:
            return _grpc(
                arr=arr, predictor_host=predictor_host, tf_data_type=tf_data_type
            )
        elif protocol == Protocol.seldon_http:
            payload = seldon.create_request(arr, seldon.SeldonPayload.NDARRAY)
            response_raw = requests.post(
                SELDON_PREDICTOR_URL_FORMAT.format(predictor_host), json=payload
            )
            if response_raw.status_code == 200:
                rh = seldon.SeldonRequestHandler(response_raw.json())
                response_list = rh.extract_request()
                return np.array(response_list)
            else:
                raise Exception(
                    "Failed to get response from model return_code"
                    ":%d" % response_raw.status_code
                )
        elif protocol == Protocol.tensorflow_http:
            instances = []
            for req_data in arr:
                if isinstance(req_data, np.ndarray):
                    instances.append(req_data.tolist())
                else:
                    instances.append(req_data)
            request = {"instances": instances}
            response = requests.post(
                TENSORFLOW_PREDICTOR_URL_FORMAT.format(predictor_host, model_name),
                json.dumps(request),
            )
            if response.status_code != 200:
                raise Exception(
                    "Failed to get response from model return_code"
                    ":%d" % response.status_code
                )
            return np.array(response.json()["predictions"])

    return _predict_fn


def _grpc(arr: np.array, predictor_host: str, tf_data_type: Optional[str]) -> np.array:
    options = [
        ("grpc.max_send_message_length", GRPC_MAX_MSG_LEN),
        ("grpc.max_receive_message_length", GRPC_MAX_MSG_LEN),
    ]
    channel = grpc.insecure_channel(predictor_host, options)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if tf_data_type is not None:
        datadef = prediction_pb2.DefaultData(
            tftensor=tf.make_tensor_proto(arr, tf_data_type)
        )
    else:
        datadef = prediction_pb2.DefaultData(tftensor=tf.make_tensor_proto(arr))
    request = prediction_pb2.SeldonMessage(data=datadef)
    response = stub.Predict(request=request)
    arr_resp = tf.make_ndarray(response.data.tftensor)
    return arr_resp
