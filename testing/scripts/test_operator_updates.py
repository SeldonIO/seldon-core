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
@pytest.mark.parametrize("from_version", ["0.4.1", "0.5.1", "1.0.0", "1.0.1"])
def test_cluster_update(namespace, from_version):
    # Install past version cluster-wide
    retry_run("helm delete seldon -n seldon-system")
    retry_run(
        "helm install seldon "
        "seldonio/seldon-core-operator "
        "--namespace seldon-system "
        f"--version {from_version} "
        "--wait"
    )

    # Deploy test model
    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # Upgrade to source code version cluster-wide.
    # Note that this upgrade should leave the cluster as it was before the
    # test.
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
        "--wait",
        attempts=2,
    )

    assert_model("mymodel", namespace, initial=True)


@pytest.mark.sequential
@pytest.mark.parametrize("from_version", ["1.0.0", "1.0.1"])
def test_namespace_update(namespace, from_version):
    # Install past version cluster-wide
    retry_run("helm delete seldon -n seldon-system")
    retry_run(
        "helm install seldon "
        "seldonio/seldon-core-operator "
        "--namespace seldon-system "
        f"--version {from_version} "
        "--wait"
    )

    # Deploy test model
    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # Label namespace to deploy a single operator
    retry_run(
        f"kubectl label namespace {namespace} seldon.io/controller-id={namespace}"
    )

    # Install on the current namespace
    retry_run(
        "helm install seldon "
        "../../helm-charts/seldon-core-operator "
        f"--namespace {namespace} "
        "--set istio.enabled=true "
        "--set istio.gateway=seldon-gateway "
        "--set certManager.enabled=false "
        "--set crd.create=false "
        "--set singleNamespace=true "
        "--wait",
        attempts=2,
    )

    # Assert that model is still working under new namespaced version
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # Delete all resources (webhooks, etc.) before deleting namespace
    retry_run(f"helm delete seldon --namespace {namespace}")

    # Re-install source code version cluster-wide
    retry_run(
        "helm upgrade seldon "
        "../../helm-charts/seldon-core-operator "
        "--namespace seldon-system "
        "--set istio.enabled=true "
        "--set istio.gateway=seldon-gateway "
        "--set certManager.enabled=false "
        "--wait",
        attempts=2,
    )
