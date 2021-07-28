import json
import logging
import time
from subprocess import run

import numpy as np
from google.protobuf import json_format

from e2e_utils import v2_protocol
from e2e_utils.models import deploy_model
from seldon_e2e_utils import (
    create_random_data,
    grpc_request_ambassador,
    initial_rest_request,
    log_sdep_logs,
    rest_request_ambassador,
    retry_run,
    wait_for_rollout,
    wait_for_status,
)


class TestPrepack(object):

    # Test prepackaged server for sklearn
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

        r = grpc_request_ambassador(
            "sklearn", namespace, data=np.array([[0.1, 0.2, 0.3, 0.4]])
        )
        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)

        logging.warning("Success for test_prepack_sklearn")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    def test_sklearn_v2(self, namespace):
        deploy_model(
            "sklearn",
            namespace=namespace,
            protocol="kfserving",
            model_implementation="SKLEARN_SERVER",
            model_uri="gs://seldon-models/sklearn/iris-0.23.2/lr_model",
        )
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(1)

        logging.warning("Initial request")
        r = v2_protocol.inference_request(
            deployment_name="sklearn",
            model_name="model",
            namespace=namespace,
            payload={
                "inputs": [
                    {
                        "name": "input-0",
                        "shape": [1, 4],
                        "datatype": "FP32",
                        "data": [[0.1, 0.2, 0.3, 0.4]],
                    }
                ]
            },
        )
        assert r.status_code == 200

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

        r = grpc_request_ambassador(
            "xgboost", namespace, data=np.array([[0.1, 0.2, 0.3, 0.4]])
        )
        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)

        logging.warning("Success for test_prepack_xgboost")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    def test_xgboost_v2(self, namespace):
        deploy_model(
            "xgboost",
            namespace=namespace,
            protocol="kfserving",
            model_implementation="XGBOOST_SERVER",
            model_uri="gs://seldon-models/xgboost/iris",
        )
        wait_for_status("xgboost", namespace)
        wait_for_rollout("xgboost", namespace)
        time.sleep(1)

        logging.warning("Initial request")
        r = v2_protocol.inference_request(
            deployment_name="xgboost",
            model_name="model",
            namespace=namespace,
            payload={
                "inputs": [
                    {
                        "name": "input-0",
                        "shape": [1, 4],
                        "datatype": "FP32",
                        "data": [[0.1, 0.2, 0.3, 0.4]],
                    }
                ]
            },
        )
        assert r.status_code == 200

    # Test prepackaged server for MLflow
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

    # test mlflow with kfserving (v2) protocol
    def test_mlflow_v2(self, namespace):
        tag = "mlflow"
        deploy_model(
            tag,
            namespace=namespace,
            protocol="kfserving",
            model_implementation="MLFLOW_SERVER",
            model_uri="gs://seldon-models/v1.10.0-dev/mlflow/elasticnet_wine",
        )
        wait_for_status(tag, namespace)
        wait_for_rollout(tag, namespace)
        time.sleep(1)

        logging.warning("Initial request")
        r = v2_protocol.inference_request(
            deployment_name=tag,
            model_name="model",
            namespace=namespace,
            payload={
                "parameters": {
                    "content_type": "pd"
                },
                "inputs": [
                    {
                      "name": "fixed acidity",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [7.4],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "volatile acidity",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [0.7000],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "citric acidity",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [0],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "residual sugar",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [1.9],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "chlorides",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [0.076],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "free sulfur dioxide",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [11],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "total sulfur dioxide",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [34],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "density",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [0.9978],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "pH",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [3.51],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "sulphates",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [0.56],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                    {
                      "name": "alcohol",
                      "shape": [1],
                      "datatype": "FP32",
                      "data": [9.4],
                      "parameters": {
                          "content_type": "np"
                      }
                    },
                ]
            },
        )
        assert r.status_code == 200

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
