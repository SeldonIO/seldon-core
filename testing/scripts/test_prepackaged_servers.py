from seldon_e2e_utils import (
    get_deployment_names,
    wait_for_rollout,
    initial_rest_request,
    retry_run,
    create_random_data,
    wait_for_status,
)
from subprocess import run
import time
import logging


class TestPrepack(object):

    # Test prepackaged server for sklearn
    def test_sklearn(self):
        namespace = "test-sklearn"
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        for deployment_name in get_deployment_names("sklearn", namespace):
            wait_for_rollout(deployment_name, namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "sklearn", namespace, data=[[0.1, 0.2, 0.3, 0.4]], dtype="ndarray"
        )
        assert r.status_code == 200
        logging.warning("Success for test_prepack_sklearn")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for tfserving
    def test_tfserving(self):
        namespace = "test-tfserving"
        spec = "../../servers/tfserving/samples/mnist_rest.yaml"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(f"kubectl apply -f {spec}  -n {namespace}")
        wait_for_status("tfserving", namespace)
        for deployment_name in get_deployment_names("tfserving", namespace):
            wait_for_rollout(deployment_name, namespace)
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
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for xgboost
    def test_xgboost(self):
        namespace = "test-xgboost"
        spec = "../../servers/xgboostserver/samples/iris.yaml"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(f"kubectl apply -f {spec}  -n {namespace}")
        wait_for_status("xgboost", namespace)
        for deployment_name in get_deployment_names("xgboost", namespace):
            wait_for_rollout(deployment_name, namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "xgboost", namespace, data=[[0.1, 0.2, 0.3, 0.4]], dtype="ndarray"
        )
        assert r.status_code == 200
        logging.warning("Success for test_prepack_xgboost")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for MLflow
    def test_mlflow(self):
        namespace = "test-mlflow"
        spec = "../../servers/mlflowserver/samples/elasticnet_wine.yaml"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("mlflow", namespace)
        for deployment_name in get_deployment_names("mlflow", namespace):
            wait_for_rollout(deployment_name, namespace)
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

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)
