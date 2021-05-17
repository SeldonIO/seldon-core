from subprocess import run

import pytest

from seldon_e2e_utils import assert_model, wait_for_rollout, wait_for_status


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
        f"--set apiVersion={apiVersion} "
        f"--set model.image=seldonio/fixed-model:0.1 "
        f"--namespace {namespace}"
    )
    run(command, shell=True, check=True)

    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)

    assert_model("mymodel", namespace, initial=True)

    run("helm delete mymodel", shell=True)
