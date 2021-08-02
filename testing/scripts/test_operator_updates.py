import pytest

from seldon_e2e_utils import (
    assert_model,
    assert_model_during_op,
    retry_run,
    wait_for_rollout,
    wait_for_status,
)

SELDON_VERSIONS_TO_TEST = ["1.0.2", "1.1.0", "1.2.3", "1.3.0", "1.4.0"]


@pytest.mark.sequential
@pytest.mark.parametrize("seldon_version", SELDON_VERSIONS_TO_TEST, indirect=True)
def test_cluster_update(namespace, seldon_version):
    # NOTE: if this fails we might need to rebuild the test (fixed) models to pick up latest python changes.
    # run:
    #   `make kind_build_test_models`
    # Deploy test model
    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    # Upgrade to source code version cluster-wide.
    def _upgrade_seldon():
        retry_run(
            "helm upgrade seldon "
            "../../helm-charts/seldon-core-operator "
            "--namespace seldon-system "
            "--wait",
            attempts=2,
        )

    assert_model_during_op(_upgrade_seldon, "mymodel", namespace)


@pytest.mark.flaky(max_runs=2)
@pytest.mark.sequential
@pytest.mark.parametrize("seldon_version", SELDON_VERSIONS_TO_TEST, indirect=True)
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

    def _install_namespace_scoped():
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

    assert_model_during_op(_install_namespace_scoped, "mymodel", namespace)


@pytest.mark.sequential
@pytest.mark.parametrize("seldon_version", SELDON_VERSIONS_TO_TEST, indirect=True)
def test_label_update(namespace, seldon_version):
    # Deploy test model
    retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True)

    controller_id = f"seldon-{namespace}"

    def _install_label_scoped():
        # TODO: We install the new controller on the same namespace
        # but it's not necessary, since it will get targeted by
        # controllerId
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

    assert_model_during_op(_install_label_scoped, "mymodel", namespace)

    # Delete all resources (webhooks, etc.) before deleting namespace
    retry_run(f"helm delete {controller_id} --namespace {namespace}")
