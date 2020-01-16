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
@pytest.mark.parametrize(
    "seldon_version",
    [
        pytest.param("0.4.1", marks=pytest.mark.skip(reason="fixed in Helm 3.0.3")),
        "0.5.1",
        "1.0.0",
        "1.0.1",
    ],
    indirect=True,
)
def test_cluster_update(namespace, seldon_version):
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
        "--wait",
        attempts=2,
    )

    assert_model("mymodel", namespace, initial=True)


@pytest.mark.sequential
@pytest.mark.parametrize("seldon_version", ["1.0.0", "1.0.1"], indirect=True)
def test_namespace_update(namespace, seldon_version):
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
        "--set crd.create=false "
        "--set singleNamespace=true "
        "--wait",
        attempts=2,
    )

    # Assert that model is still working under new namespaced version
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)


@pytest.mark.sequential
@pytest.mark.parametrize("seldon_version", ["1.0.0", "1.0.1"], indirect=True)
def test_label_update(namespace, seldon_version):
    # Deploy test model
    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # Install id-scoped operator
    controller_id = f"seldon-{namespace}"
    # TODO: We install the new controller on the same namespace but it's not
    # necessary, since it will get targeted by controllerId
    retry_run(
        f"helm install {controller_id} "
        "../../helm-charts/seldon-core-operator "
        f"--namespace {namespace} "
        "--set crd.create=false "
        f"--set controllerId={controller_id} "
        "--wait",
        attempts=2,
    )

    # Label model to be served by new controller
    retry_run(
        "kubectl label sdep mymodel "
        f"seldon.io/controller-id={controller_id} "
        f"--namespace {namespace}"
    )

    # Assert that model is still working under new id-scoped operator
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # Delete all resources (webhooks, etc.) before deleting namespace
    retry_run(f"helm delete {controller_id} --namespace {namespace}")
