import time
import logging
import pytest
from subprocess import run

from seldon_e2e_utils import (
    wait_for_rollout,
    initial_rest_request,
    rest_request_ambassador,
    retry_run,
    create_random_data,
    wait_for_status,
    rest_request,
    log_sdep_logs,
)
from e2e_utils import v2_protocol
from conftest import SELDON_E2E_TESTS_USE_EXECUTOR

skipif_engine = pytest.mark.skipif(
    not SELDON_E2E_TESTS_USE_EXECUTOR, reason="Not supported by the Java engine"
)


class TestPrepack(object):

    # Test prepackaged server for sklearn
    @skipif_engine
    def test_sklearn(self, namespace):
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "sklearn", namespace, data=[[0.1, 0.2, 0.3, 0.4]], dtype="ndarray"
        )
        assert r.status_code == 200

        r = rest_request_ambassador("sklearn", namespace, method="metadata")
        assert r.status_code == 200

        res = r.json()
        logging.warning(res)
        assert res["name"] == "iris"
        assert res["versions"] == ["iris/v1"]

        logging.warning("Success for test_prepack_sklearn")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    @skipif_engine
    def test_sklearn_v2(self, namespace):
        spec = "../resources/iris-sklearn-v2.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(1)

        logging.warning("Initial request")
        r = v2_protocol.inference_request(
            model_name="sklearn",
            namespace=namespace,
            payload={
                "inputs": [
                    {
                        "name": "input-0",
                        "shape": [1, 4],
                        "datatype": "FP32",
                        "data": [[0.1, 0.2, 0.3, 0.4]],
                    }
                ],
            },
        )
        assert r.status_code == 200
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    # Test prepackaged server for tfserving
    def test_tfserving(self, namespace):
        spec = "../../servers/tfserving/samples/mnist_rest.yaml"
        retry_run(f"kubectl apply -f {spec}  -n {namespace}")
        wait_for_status("tfserving", namespace)
        wait_for_rollout("tfserving", namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "tfserving",
            namespace,
            data=[create_random_data(784)[1].tolist()],
            dtype="ndarray",
        )
        assert r.status_code == 200
        logging.warning("Success for test_prepack_tfserving")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    # Test prepackaged server for xgboost
    @skipif_engine
    def test_xgboost(self, namespace):
        spec = "../../servers/xgboostserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec}  -n {namespace}")
        wait_for_status("xgboost", namespace)
        wait_for_rollout("xgboost", namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "xgboost", namespace, data=[[0.1, 0.2, 0.3, 0.4]], dtype="ndarray"
        )
        assert r.status_code == 200

        r = rest_request_ambassador("xgboost", namespace, method="metadata")
        assert r.status_code == 200

        res = r.json()
        logging.warning(res)
        assert res["name"] == "xgboost-iris"
        assert res["versions"] == ["xgboost-iris/v1"]

        logging.warning("Success for test_prepack_xgboost")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    @skipif_engine
    def test_xgboost_v2(self, namespace):
        spec = "../resources/iris-xgboost-v2.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("xgboost", namespace)
        wait_for_rollout("xgboost", namespace)
        time.sleep(1)

        logging.warning("Initial request")
        r = v2_protocol.inference_request(
            model_name="xgboost",
            namespace=namespace,
            payload={
                "inputs": [
                    {
                        "name": "input-0",
                        "shape": [1, 4],
                        "datatype": "FP32",
                        "data": [[0.1, 0.2, 0.3, 0.4]],
                    }
                ],
            },
        )
        assert r.status_code == 200
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    # Test prepackaged server for MLflow
    @skipif_engine
    def test_mlflow(self, namespace):
        spec = "../../servers/mlflowserver/samples/elasticnet_wine.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("mlflow", namespace)
        wait_for_rollout("mlflow", namespace)
        time.sleep(1)

        r = initial_rest_request(
            "mlflow",
            namespace,
            data=[[6.3, 0.3, 0.34, 1.6, 0.049, 14, 132, 0.994, 3.3, 0.49, 9.5]],
            dtype="ndarray",
            names=[
                "fixed acidity",
                "volatile acidity",
                "citric acid",
                "residual sugar",
                "chlorides",
                "free sulfur dioxide",
                "total sulfur dioxide",
                "density",
                "pH",
                "sulphates",
                "alcohol",
            ],
        )
        assert r.status_code == 200

        r = rest_request_ambassador("mlflow", namespace, method="metadata")
        assert r.status_code == 200

        res = r.json()
        logging.warning(res)
        assert res["name"] == "mlflow-wines"
        assert res["versions"] == ["mlflow-wines/v1"]

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    # Test prepackaged Text SKLearn Alibi Explainer
    def test_text_alibi_explainer(self, namespace):
        spec = "../resources/movies-text-explainer.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("movie", namespace)
        wait_for_rollout("movie", namespace, expected_deployments=2)
        time.sleep(5)
        logging.warning("Initial request")
        r = initial_rest_request(
            "movie", namespace, data=["This is test data"], dtype="ndarray"
        )
        log_sdep_logs("movie", namespace)
        assert r.status_code == 200

        # First request most likely will fail because AnchorText explainer
        # is creating the explainer on first request - we skip checking output
        # of it, sleep for some time and then do the actual explanation request
        # we use in the test
        e = initial_rest_request(
            "movie",
            namespace,
            data=["This is test data"],
            dtype="ndarray",
            method="explain",
            predictor_name="movies-predictor",
        )
        log_sdep_logs("movie", namespace)

        time.sleep(30)

        e = initial_rest_request(
            "movie",
            namespace,
            data=["This is test data"],
            dtype="ndarray",
            method="explain",
            predictor_name="movies-predictor",
        )
        log_sdep_logs("movie", namespace)
        assert e.status_code == 200
        logging.warning("Success for test_prepack_sklearn")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    # Test openAPI endpoints for documentation
    @skipif_engine
    def test_openapi_sklearn(self, namespace):
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(1)
        logging.warning("Initial request")

        r = initial_rest_request("sklearn", namespace, method="openapi_ui")
        assert r.status_code == 200
        content_type_header = r.headers.get("content-type")
        assert "text/html" in content_type_header

        r = initial_rest_request("sklearn", namespace, method="openapi_schema")
        assert r.status_code == 200
        openapi_schema = r.json()
        assert "openapi" in openapi_schema

        logging.warning("Success for test_openapi_sklearn")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
