import logging
from subprocess import run

from seldon_e2e_utils import (
    API_AMBASSADOR,
    initial_rest_request,
    rest_request_ambassador,
    retry_run,
    wait_for_rollout,
    wait_for_status,
)


class TestLocalOperators(object):
    def test_namespace_operator(self, namespace):
        retry_run(
            f"helm install seldon ../../helm-charts/seldon-core-operator --namespace {namespace} --set istio.enabled=true --set istio.gateway=istio-system/seldon-gateway --set certManager.enabled=false --set crd.create=false --set singleNamespace=true"
        )
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace)
        logging.warning("Initial request")
        r = initial_rest_request("mymodel", namespace, endpoint=API_AMBASSADOR)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        logging.warning("Success for test_namespace_operator")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"helm uninstall seldon -n {namespace}", shell=True)

    def test_labelled_operator(self, namespace):
        retry_run(
            f"helm install seldon ../../helm-charts/seldon-core-operator --namespace {namespace} --set istio.enabled=true --set istio.gateway=istio-system/seldon-gateway --set certManager.enabled=false --set crd.create=false --set controllerId=seldon-id1"
        )
        retry_run(
            f"kubectl apply -f ../resources/model_controller_id.yaml -n {namespace}"
        )
        wait_for_status("test-c1", namespace)
        wait_for_rollout("test-c1", namespace)
        logging.warning("Initial request")
        r = initial_rest_request("test-c1", namespace, endpoint=API_AMBASSADOR)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        logging.warning("Success for test_labelled_operator")
        run(
            f"kubectl delete -f ../resources/model_controller_id.yaml -n {namespace}",
            shell=True,
        )
        run(f"helm uninstall seldon -n {namespace}", shell=True)
