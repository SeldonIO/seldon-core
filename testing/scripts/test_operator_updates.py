import pytest

from seldon_e2e_utils import (
    initial_rest_request,
    rest_request,
    retry_run,
    wait_for_status,
    wait_for_rollout,
)


@pytest.mark.serial
@pytest.mark.parametrize("from_version", ["0.4.1", "0.5.1", "1.0.0"])
def test_operator_update(namespace, from_version):
    retry_run("helm delete seldon -n seldon-system")
    retry_run(
        "helm install seldon "
        "seldonio/seldon-core-operator "
        "--namespace seldon-system "
        f"--version {from_version} "
    )

    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)

    r = initial_rest_request("mymodel", namespace)
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]

    # The upgrade should leave the cluster as it was before the test
    retry_run(
        "helm upgrade seldon "
        "../../helm-charts/seldon-core-operator "
        "--namespace seldon-system "
        "--set istio.enabled=true "
        "--set istio.gateway=seldon-gateway "
        "--set certManager.enabled=false"
    )

    # Improve health check and move to func (e.g. also checking list of sdep)
    r = rest_request("mymodel", namespace)
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
