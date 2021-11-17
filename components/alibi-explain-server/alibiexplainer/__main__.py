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
# Copied from https://github.com/kubeflow/kfserving/blob/master/python/alibiexplainer
# /alibiexplainer/__main__.py
# and modified since
#

import logging
import os
import sys

import kfserving
from alibi.saving import load_explainer
from tensorflow import keras

from alibiexplainer import AlibiExplainer
from alibiexplainer.constants import SELDON_LOGLEVEL
from alibiexplainer.parser import parse_args
from alibiexplainer.server import ExplainerServer
from alibiexplainer.utils import (  # pylint:disable=no-name-in-module
    ExplainerMethod, Protocol, construct_predict_fn)

logging.basicConfig(level=SELDON_LOGLEVEL)
KERAS_MODEL = "model.h5"


def main():
    args, extra = parse_args(sys.argv[1:])
    # Pretrained Alibi explainer
    alibi_model = None
    keras_model = None
    predict_fn = construct_predict_fn(
        predictor_host=args.predictor_host,
        model_name=args.model_name,
        protocol=Protocol(args.protocol),
        tf_data_type=args.tf_data_type,
    )
    if args.storage_uri is not None:
        path = kfserving.Storage.download(args.storage_uri)
        if os.path.exists(path):
            logging.info("Loading Alibi model")
            alibi_model = load_explainer(predictor=predict_fn, path=path)
        else:
            keras_path = os.path.join(path, KERAS_MODEL)
            if os.path.exists(keras_path):
                with open(keras_path, "rb") as f:
                    logging.info("Loading Keras model")
                    keras_model = keras.models.load_model(keras_path)

    explainer = AlibiExplainer(
        name=args.model_name,
        predict_fn=predict_fn,
        method=ExplainerMethod(args.command),
        config=extra,
        explainer=alibi_model,
        protocol=Protocol(args.protocol),
        keras_model=keras_model,
    )
    explainer.load()
    ExplainerServer(args.http_port).start(explainer)


if __name__ == "__main__":
    main()
