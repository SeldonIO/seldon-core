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
import sys

from alibiexplainer import AlibiExplainer
from alibiexplainer.constants import SELDON_LOGLEVEL
from alibiexplainer.parser import parse_args
from alibiexplainer.server import ExplainerServer
from alibiexplainer.utils import (
    ExplainerMethod,
    Protocol,
    construct_predict_fn,
    get_persisted_explainer,
    get_persisted_keras,
    is_persisted_explainer,
    is_persisted_keras,
)

logging.basicConfig(level=SELDON_LOGLEVEL)


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
        # we assume here that model is local
        path = args.storage_uri

        if is_persisted_explainer(path):
            alibi_model = get_persisted_explainer(predict_fn=predict_fn, dirname=path)

        if is_persisted_keras(path):
            keras_model = get_persisted_keras(path)

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
