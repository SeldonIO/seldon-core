import logging
import time
from subprocess import run

import pytest
from seldon_e2e_utils import (API_AMBASSADOR, API_ISTIO_GATEWAY, assert_model,
                              assert_model_during_op, get_pod_name_for_sdep,
                              initial_rest_request, rest_request_ambassador,
                              retry_run, to_resources_path,
                              wait_for_pod_shutdown, wait_for_rollout,
                              wait_for_status)

with_api_gateways = pytest.mark.parametrize(
    "api_gateway", [API_AMBASSADOR, API_ISTIO_GATEWAY], ids=["ambas", "istio"]
)


@pytest.mark.sequential
@pytest.mark.flaky(max_runs=5)
@with_api_gateways
class TestRollingHttp(object):
    # Test updating a model to a multi predictor model
    def test_rolling_update5(self, namespace, api_gateway):
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace)
        logging.warning("Initial request")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph6.json -n {namespace}")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", namespace, api_gateway)
            assert r.status_code == 200
            res = r.json()
            assert (res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            )
            if (not r.status_code == 200) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            ):
                break
            time.sleep(1)
        assert i < 100
        logging.warning("Success for test_rolling_update5")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph6.json -n {namespace}", shell=True)

    # Test updating a model with a new image version as the only change
    def test_rolling_update6(self, namespace, api_gateway):
        retry_run(f"kubectl apply -f ../resources/graph1svc.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace, expected_deployments=2)
        logging.warning("Initial request")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph2svc.json -n {namespace}")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", namespace, api_gateway)
            assert r.status_code == 200
            res = r.json()
            assert (res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            )
            if (not r.status_code == 200) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            ):
                break
            time.sleep(1)
        assert i < 100
        logging.warning("Success for test_rolling_update6")
        run(f"kubectl delete -f ../resources/graph1svc.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph2svc.json -n {namespace}", shell=True)

    # test changing the image version and the name of its container
    def test_rolling_update7(self, namespace, api_gateway):
        retry_run(f"kubectl apply -f ../resources/graph1svc.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace, expected_deployments=2)
        logging.warning("Initial request")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph3svc.json -n {namespace}")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", namespace, api_gateway)
            assert r.status_code == 200
            res = r.json()
            assert (res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            )
            if (not r.status_code == 200) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            ):
                break
            time.sleep(1)
        assert i < 100
        logging.warning("Success for test_rolling_update7")
        run(f"kubectl delete -f ../resources/graph1svc.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph3svc.json -n {namespace}", shell=True)

    # Test updating a model with a new resource request but same image
    def test_rolling_update8(self, namespace, api_gateway):
        retry_run(f"kubectl apply -f ../resources/graph1svc.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace, expected_deployments=2)
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph4svc.json -n {namespace}")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", namespace, api_gateway)
            assert r.status_code == 200
            res = r.json()
            assert res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
            time.sleep(1)
        assert i == 49
        logging.warning("Success for test_rolling_update8")
        run(f"kubectl delete -f ../resources/graph1svc.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph4svc.json -n {namespace}", shell=True)

    # Test updating a model with a multi deployment new model
    def test_rolling_update9(self, namespace, api_gateway):
        retry_run(f"kubectl apply -f ../resources/graph1svc.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace, expected_deployments=2)
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph5svc.json -n {namespace}")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", namespace, api_gateway)
            assert r.status_code == 200
            res = r.json()
            assert res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
            time.sleep(1)
        assert i == 49
        logging.warning("Success for test_rolling_update9")
        run(f"kubectl delete -f ../resources/graph1svc.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph5svc.json -n {namespace}", shell=True)

    # Test updating a model to a multi predictor model
    def test_rolling_update10(self, namespace, api_gateway):
        retry_run(f"kubectl apply -f ../resources/graph1svc.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace, expected_deployments=2)
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph6svc.json -n {namespace}")
        r = initial_rest_request("mymodel", namespace, endpoint=api_gateway)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", namespace, api_gateway)
            assert r.status_code == 200
            res = r.json()
            assert (res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            )
            if (not r.status_code == 200) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            ):
                break
            time.sleep(1)
        assert i < 100
        logging.warning("Success for test_rolling_update10")
        run(f"kubectl delete -f ../resources/graph1svc.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph6svc.json -n {namespace}", shell=True)


@pytest.mark.flaky(max_runs=3)
@with_api_gateways
@pytest.mark.parametrize(
    "from_deployment,to_deployment,change",
    [
        ("graph1.json", "graph2.json", True),  # New image version
        (
            "graph1.json",
            "graph3.json",
            True,
        ),  # New image version and new name of container
        ("graph1.json", "graph4.json", True),  # New resource request but same image
        ("graph1.json", "graph5.json", True),  # Update with multi-deployment new model
        ("graph1.json", "graph8.json", True),  # From v1alpha2 to v1
        ("graph7.json", "graph8.json", False),  # From v1alpha3 to v1
    ],
)
def test_rolling_deployment(
    namespace, api_gateway, from_deployment, to_deployment, change
):
    from_file_path = to_resources_path(from_deployment)
    retry_run(f"kubectl apply -f {from_file_path} -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    assert_model("mymodel", namespace, initial=True, endpoint=api_gateway)

    old_pod_name = get_pod_name_for_sdep("mymodel", namespace)[0]
    to_file_path = to_resources_path(to_deployment)

    def _update_model():
        retry_run(f"kubectl apply -f {to_file_path} -n {namespace}")
        if change:
            wait_for_pod_shutdown(old_pod_name, namespace)
        wait_for_status("mymodel", namespace)
        time.sleep(2)  # Wait a little after deployment marked Available

    assert_model_during_op(_update_model, "mymodel", namespace, endpoint=api_gateway)

    delete_cmd = f"kubectl delete --ignore-not-found -n {namespace}"
    run(f"{delete_cmd} -f {from_file_path}", shell=True)
    run(f"{delete_cmd} -f {to_file_path}", shell=True)
