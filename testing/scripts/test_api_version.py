import pytest
from seldon_e2e_utils import (
    wait_for_rollout,
    wait_for_status,
    initial_rest_request,
    rest_request_ambassador,
    API_AMBASSADOR,
)
from subprocess import run


@pytest.mark.parametrize(
    "apiVersion",
    [
        "machinelearning.seldon.io/v1alpha2",
        "machinelearning.seldon.io/v1alpha3",
        "machinelearning.seldon.io/v1",
    ],
)
def test_api_version(namespace, apiVersion):
    command = (
        "helm install mymodel ../../helm-charts/seldon-single-model "
        "--set oauth.key=oauth-key "
        "--set oauth.secret=oauth-secret "
        f"--set apiVersion={apiVersion} "
        f"--namespace {namespace}"
    )
    run(command, shell=True, check=True)

    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    initial_rest_request("mymodel", namespace)

    r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)

    assert r.status_code == 200
    assert len(r.json()["data"]["tensor"]["values"]) == 1

    run(f"helm delete mymodel", shell=True)
