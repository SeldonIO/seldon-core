import json
import time
from subprocess import run

import numpy as np
import pytest
import requests
from tenacity import Retrying, stop_after_attempt, wait_fixed

from seldon_e2e_utils import retry_run, wait_for_deployment

# NOTE:
# to recreate the artifacts for these test:
# 1. use notebooks/explainer_examples_v2.ipynb to create them
# 2. upload to gs

AFTER_WAIT_SLEEP = 20
TENACITY_WAIT = 10
TENACITY_STOP_AFTER_ATTEMPT = 5


class TestExplainV2Server:
    @pytest.mark.sequential
    def test_alibi_explain_anchor_tabular(self, namespace):
        spec = "../resources/iris_anchor_tabular_explainer_v2.yaml"
        name = "iris-default-explainer"
        vs_prefix = (
            f"seldon/{namespace}/iris-explainer/default/v2/models/"
            f"iris-default-explainer/infer"
        )

        test_data = np.array([[5.964, 4.006, 2.081, 1.031]])
        inference_request = {
            "parameters": {"content_type": "np"},
            "inputs": [
                {
                    "name": "explain",
                    "shape": test_data.shape,
                    "datatype": "FP32",
                    "data": test_data.tolist(),
                    "parameters": {"content_type": "np"},
                },
            ],
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
                # note: explanation will come back in v2 as a nested json dictionary
                explanation = json.loads(r.json()["outputs"][0]["data"])

        assert explanation["meta"]["name"] == "AnchorTabular"
        assert "anchor" in explanation["data"]
        assert "precision" in explanation["data"]
        assert "coverage" in explanation["data"]

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
