import os
import time
import pytest

from e2e_utils import v2_protocol
from e2e_utils.models import deploy_model
from seldon_e2e_utils import retry_run, wait_for_rollout, wait_for_status

from conftest import RESOURCES_PATH


#  @pytest.mark.parametrize("use_grpc", [True, False])
@pytest.mark.parametrize("use_grpc", [True])
def test_graph_v2(namespace: str, use_grpc: bool):
    sdep_name = "graph-test"
    spec = os.path.join(RESOURCES_PATH, "graph-v2.yaml")
    retry_run(f"kubectl apply -f {spec}  -n {namespace}")
    wait_for_status(sdep_name, namespace)
    wait_for_rollout(sdep_name, namespace)
    time.sleep(1)

    inference_request = (
        v2_protocol.inference_request_grpc
        if use_grpc
        else v2_protocol.inference_request
    )

    response = inference_request(
        deployment_name=sdep_name,
        namespace=namespace,
        payload={"inputs": []},
    )
    assert len(response["outputs"]) == 1
