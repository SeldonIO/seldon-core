import time
from subprocess import run

import numpy as np
import pytest
import requests
from tenacity import Retrying, stop_after_attempt, wait_fixed

from seldon_e2e_utils import retry_run, wait_for_deployment

# NOTE:
# to recreate the artifacts for these test:
# anchor tabular:
# 1. python components/alibi-explain-server/tests/make_test_models.py
# --model anchor_tabular
# --model_dir <dir_name>
# 2. upload <dir_name> to gs
# 3. patch ../resources/iris_anchor_tabular_explainer.yaml to reflect change
#
# anchor image:
# 1. python components/alibi-explain-server/tests/make_test_models.py
# --model anchor_image
# --model_dir <dir_name>
# 2. upload <dir_name> to gs
# 3. patch ../resources/tf_cifar_anchor_image_explainer.yaml to reflect change

# note: create models using poetry env for explainer v1 server

AFTER_WAIT_SLEEP = 20
TENACITY_WAIT = 10
TENACITY_STOP_AFTER_ATTEMPT = 5


class TestExplainServer:
    @pytest.mark.sequential
    def test_alibi_explain_anchor_tabular(self, namespace):
        spec = "../resources/iris_anchor_tabular_explainer.yaml"
        name = "iris-default-explainer"
        vs_prefix = f"seldon/{namespace}/iris-explainer/default/api/v1.0/explain"

        inference_request = {
            "data": {
                "names": ["text"],
                "ndarray": [[5.964, 4.006, 2.081, 1.031]],
            }
        }

        retry_run(f"kubectl apply -f {spec} -n {namespace}")

        wait_for_deployment(name, namespace)

        time.sleep(AFTER_WAIT_SLEEP)

        for attempt in Retrying(
            wait=wait_fixed(TENACITY_WAIT),
            stop=stop_after_attempt(TENACITY_STOP_AFTER_ATTEMPT),
        ):
            with attempt:
                r = requests.post(
                    f"http://localhost:8004/{vs_prefix}",
                    json=inference_request,
                )
                explanation = r.json()

        assert explanation["meta"]["name"] == "AnchorTabular"
        assert "anchor" in explanation["data"]
        assert "precision" in explanation["data"]
        assert "coverage" in explanation["data"]

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    @pytest.mark.sequential
    def test_alibi_explain_anchor_image_tensorflow_protocol(self, namespace):
        spec = "../resources/tf_cifar_anchor_image_explainer.yaml"
        name = "cifar10-default-explainer"
        vs_prefix = (
            f"seldon/{namespace}/cifar10-explainer/default/v1/models/"
            f"cifar10-classifier:explain"
        )
        retry_run(f"kubectl apply -f {spec} -n {namespace}")

        wait_for_deployment(name, namespace)

        test_data = np.random.randn(32, 32, 3)
        inference_request = {"instances": f"{test_data.tolist()}"}

        for attempt in Retrying(
            wait=wait_fixed(TENACITY_WAIT),
            stop=stop_after_attempt(TENACITY_STOP_AFTER_ATTEMPT),
        ):
            with attempt:
                r = requests.post(
                    f"http://localhost:8004/{vs_prefix}",
                    json=inference_request,
                )
                explanation = r.json()

        assert explanation["meta"]["name"] == "AnchorImage"
        assert "anchor" in explanation["data"]
        assert "precision" in explanation["data"]
        assert "coverage" in explanation["data"]

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
