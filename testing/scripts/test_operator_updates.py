import pytest

from seldon_e2e_utils import (
    initial_rest_request,
    rest_request,
    retry_run,
    wait_for_status,
    wait_for_rollout,
)


def assert_model(sdep_name, namespace, initial=False):
    _request = initial_rest_request if initial else rest_request
    r = _request(sdep_name, namespace)

    assert r is not None
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]

    retry_run(f"kubectl get -n {namespace} sdep {sdep_name}")


@pytest.mark.sequential
@pytest.mark.parametrize("from_version", ["0.4.1", "0.5.1", "1.0.0"])
def test_cluster_update(namespace, from_version):
    retry_run("helm delete seldon -n seldon-system")
    retry_run(
        "helm install seldon "
        "seldonio/seldon-core-operator "
        "--namespace seldon-system "
        f"--version {from_version} "
        "--wait"
    )

    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # The upgrade should leave the cluster as it was before the test
    # TODO: There is currently a bug updating from 0.4.1 using Helm 3.0.2.
    # This will be fixed once https://github.com/helm/helm/pull/7269 is
    # released in Helm 3.0.3.
    retry_run(
        "helm upgrade seldon "
        "../../helm-charts/seldon-core-operator "
        "--namespace seldon-system "
        "--set istio.enabled=true "
        "--set istio.gateway=seldon-gateway "
        "--set certManager.enabled=false "
        "--wait"
    )

    assert_model("mymodel", namespace, initial=True)
