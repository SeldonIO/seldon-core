import requests
from requests.auth import HTTPBasicAuth
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import grpc
import numpy as np
import time
from subprocess import run, Popen
import subprocess
import json
from retrying import retry
import logging

API_AMBASSADOR = "localhost:8003"
API_ISTIO_GATEWAY = "localhost:8004"


def get_s2i_python_version():
    ret = Popen(
        "cd ../../wrappers/s2i/python && grep 'IMAGE_VERSION=' Makefile | cut -d'=' -f2",
        shell=True,
        stdout=subprocess.PIPE,
    )
    output = ret.stdout.readline()
    version = output.decode("utf-8").rstrip()
    return version


def get_seldon_version():
    ret = Popen(
        "cat ../../version.txt", shell=True, stdout=subprocess.PIPE
    )
    output = ret.stdout.readline()
    version = output.decode("utf-8").strip()
    return version


def wait_for_shutdown(deployment_name, namespace):
    ret = run(f"kubectl get -n {namespace} deploy/{deployment_name}", shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run(f"kubectl get -n {namespace} deploy/{deployment_name}", shell=True)


def get_deployment_names(sdep_name, namespace, attempts=20, sleep=5):
    for _ in range(attempts):
        ret = run(
            f"kubectl get -n {namespace} sdep {sdep_name} -o json",
            shell=True,
            capture_output=True,
        )
        if ret.returncode == 0:
            logging.warning(
                f"Successfully waited for SeldonDeployment {sdep_name}"
            )
            break
        logging.warning(
            f"Unsuccessful wait command but retrying for SeldonDeployment {sdep_name}"
        )
        time.sleep(sleep)
    assert ret.returncode == 0, "Failed to get deployment names: non-zero return code"
    data = json.loads(ret.stdout)
    # The `deploymentStatus` is dictionary which keys are names of deployments
    return list(data["status"]["deploymentStatus"])


def wait_for_rollout(deployment_name, namespace, attempts=20, sleep=5):
    for attempts in range(attempts):
        ret = run(
            f"kubectl rollout status -n {namespace} deploy/{deployment_name}", shell=True
        )
        if ret.returncode == 0:
            logging.warning(f"Successfully waited for deployment {deployment_name}")
            break
        logging.warning(f"Unsuccessful wait command but retrying for {deployment_name}")
        time.sleep(sleep)
    assert ret.returncode == 0, "Wait for rollout failed: non-zero return code"


def retry_run(cmd, attempts=10, sleep=5):
    for i in range(attempts):
        ret = run(cmd, shell=True)
        if ret.returncode == 0:
            logging.warning(f"Successfully ran command: {cmd}")
            break
        logging.warning(f"Unsuccessful command but retrying: {cmd}")
        time.sleep(sleep)
    assert ret.returncode == 0, f"Non-zero return code in retry_run for {cmd}"


def wait_for_status(name, namespace, attempts=20, sleep=5):
    for _ in range(attempts):
        ret = run(
            f"kubectl get sdep {name} -n {namespace} -o json",
            shell=True,
            check=True,
            stdout=subprocess.PIPE,
        )
        data = json.loads(ret.stdout)
        if ("status" in data) and ("deploymentStatus" in data["status"]):
            return data
        else:
            logging.warning("Failed to find status - sleeping")
            time.sleep(sleep)


def rest_request(
    model,
    namespace,
    endpoint=API_AMBASSADOR,
    data_size=5,
    rows=1,
    data=None,
    dtype="tensor",
    names=None,
):
    try:
        r = rest_request_ambassador(
            model,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
            dtype=dtype,
            names=names,
        )
        if not r.status_code == 200:
            logging.warning(f"Bad status:{r.status_code}")
            return None
        else:
            return r
    except Exception as e:
        logging.warning(f"Failed on REST request {str(e)}")
        return None


def initial_rest_request(
    model,
    namespace,
    endpoint=API_AMBASSADOR,
    data_size=5,
    rows=1,
    data=None,
    dtype="tensor",
    names=None,
):
    r = rest_request(
        model,
        namespace,
        endpoint=endpoint,
        data_size=data_size,
        rows=rows,
        data=data,
        dtype=dtype,
        names=names,
    )
    if r is None or r.status_code != 200:
        logging.warning("Sleeping 1 sec and trying again")
        time.sleep(1)
        r = rest_request(
            model,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
            dtype=dtype,
            names=names,
        )
        if r is None or r.status_code != 200:
            logging.warning("Sleeping 5 sec and trying again")
            time.sleep(5)
            r = rest_request(
                model,
                namespace,
                endpoint=endpoint,
                data_size=data_size,
                rows=rows,
                data=data,
                dtype=dtype,
                names=names,
            )
            if r is None or r.status_code != 200:
                logging.warning("Sleeping 10 sec and trying again")
                time.sleep(10)
                r = rest_request(
                    model,
                    namespace,
                    endpoint=endpoint,
                    data_size=data_size,
                    rows=rows,
                    data=data,
                    dtype=dtype,
                    names=names,
                )
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
    deployment_name,
    namespace,
    endpoint=API_AMBASSADOR,
    data_size=5,
    rows=1,
    data=None,
    dtype="tensor",
    names=None,
):
    if data is None:
        shape, arr = create_random_data(data_size, rows)
    elif dtype == "tensor":
        shape = data.shape
        arr = data.flatten()
    else:
        arr = data

    if dtype == "tensor":
        payload = {"data": {"tensor": {"shape": shape, "values": arr.tolist()}}}
    else:
        payload = {"data": {"ndarray": arr}}

    if names is not None:
        payload["data"]["names"] = names

    if namespace is None:
        response = requests.post(
            "http://"
            + endpoint
            + "/seldon/"
            + deployment_name
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
            + deployment_name
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
    deployment_name,
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
            + deployment_name
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
            + deployment_name
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
    deployment_name, namespace, endpoint="localhost:8004", data_size=5, rows=1, data=None
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
        metadata = [("seldon", deployment_name)]
    else:
        metadata = [("seldon", deployment_name), ("namespace", namespace)]
    response = stub.Predict(request=request, metadata=metadata)
    return response


def grpc_request_ambassador2(
    deployment_name, namespace, endpoint="localhost:8004", data_size=5, rows=1, data=None
):
    try:
        return grpc_request_ambassador(
            deployment_name,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
        )
    except:
        logging.warning("Warning - caught exception")
        return grpc_request_ambassador(
            deployment_name,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
        )
