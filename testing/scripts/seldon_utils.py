import requests
from requests.auth import HTTPBasicAuth
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import grpc
import numpy as np
from k8s_utils import *


def wait_for_rollout(deploymentName, namespace):
    ret = run(
        f"kubectl rollout status -n {namespace} deploy/" + deploymentName, shell=True
    )
    while ret.returncode > 0:
        time.sleep(1)
        ret = run(
            f"kubectl rollout status -n {namespace} deploy/" + deploymentName,
            shell=True,
        )


def rest_request(model, namespace):
    try:
        r = rest_request_ambassador(model, namespace, API_AMBASSADOR)
        if not r.status_code == 200:
            print("Bad status:", r.status_code)
            return None
        else:
            return r
    except Exception as e:
        print("Failed on REST request ", str(e))
        return None


def initial_rest_request(model, namespace):
    r = rest_request(model, namespace)
    if r is None:
        print("Sleeping 1 sec and trying again")
        time.sleep(1)
        r = rest_request(model, namespace)
        if r is None:
            print("Sleeping 5 sec and trying again")
            time.sleep(5)
            r = rest_request(model, namespace)
            if r is None:
                print("Sleeping 10 sec and trying again")
                time.sleep(10)
                r = rest_request(model, namespace)
    return r


def create_random_data(data_size, rows=1):
    shape = [rows, data_size]
    arr = np.random.rand(rows * data_size)
    return (shape, arr)


@retry(
    wait_exponential_multiplier=1000,
    wait_exponential_max=10000,
    stop_max_attempt_number=5,
)
def rest_request_ambassador(
    deploymentName, namespace, endpoint="localhost:8003", data_size=5, rows=1, data=None
):
    if data is None:
        shape, arr = create_random_data(data_size, rows)
    else:
        shape = data.shape
        arr = data.flatten()
    payload = {
        "data": {
            "names": ["a", "b"],
            "tensor": {"shape": shape, "values": arr.tolist()},
        }
    }
    if namespace is None:
        response = requests.post(
            "http://"
            + endpoint
            + "/seldon/"
            + deploymentName
            + "/api/v0.1/predictions",
            json=payload,
        )
    else:
        response = requests.post(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deploymentName
            + "/api/v0.1/predictions",
            json=payload,
        )
    return response


@retry(
    wait_exponential_multiplier=1000,
    wait_exponential_max=10000,
    stop_max_attempt_number=5,
)
def rest_request_ambassador_auth(
    deploymentName,
    namespace,
    username,
    password,
    endpoint="localhost:8003",
    data_size=5,
    rows=1,
    data=None,
):
    if data is None:
        shape, arr = create_random_data(data_size, rows)
    else:
        shape = data.shape
        arr = data.flatten()
    payload = {
        "data": {
            "names": ["a", "b"],
            "tensor": {"shape": shape, "values": arr.tolist()},
        }
    }
    if namespace is None:
        response = requests.post(
            "http://"
            + endpoint
            + "/seldon/"
            + deploymentName
            + "/api/v0.1/predictions",
            json=payload,
            auth=HTTPBasicAuth(username, password),
        )
    else:
        response = requests.post(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deploymentName
            + "/api/v0.1/predictions",
            json=payload,
            auth=HTTPBasicAuth(username, password),
        )
    return response


@retry(
    wait_exponential_multiplier=1000,
    wait_exponential_max=10000,
    stop_max_attempt_number=5,
)
def grpc_request_ambassador(
    deploymentName, namespace, endpoint="localhost:8004", data_size=5, rows=1, data=None
):
    if data is None:
        shape, arr = create_random_data(data_size, rows)
    else:
        shape = data.shape
        arr = data.flatten()
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=shape, values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [("seldon", deploymentName)]
    else:
        metadata = [("seldon", deploymentName), ("namespace", namespace)]
    response = stub.Predict(request=request, metadata=metadata)
    return response


def grpc_request_ambassador2(
    deploymentName, namespace, endpoint="localhost:8004", data_size=5, rows=1, data=None
):
    try:
        return grpc_request_ambassador(
            deploymentName,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
        )
    except:
        print("Warning - caught exception")
        return grpc_request_ambassador(
            deploymentName,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
        )
