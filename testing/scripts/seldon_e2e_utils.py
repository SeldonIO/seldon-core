import json
import logging
import os
import re
import subprocess
import time
from concurrent.futures import ThreadPoolExecutor, wait
from subprocess import CalledProcessError, Popen, run

import grpc
import numpy as np
import requests
from google.protobuf import empty_pb2
from requests.auth import HTTPBasicAuth
from tenacity import retry, stop_after_attempt, wait_exponential
import pandas as pd

from seldon_core.proto import prediction_pb2, prediction_pb2_grpc

API_AMBASSADOR = "localhost:8003"
API_ISTIO_GATEWAY = "localhost:8004"

TESTING_ROOT_PATH = os.path.dirname(os.path.dirname(__file__))
RESOURCES_PATH = os.path.join(TESTING_ROOT_PATH, "resources")

BENCHMARK_PARALLELISM = 2


def get_seldon_version():
    ret = Popen("cat ../../version.txt", shell=True, stdout=subprocess.PIPE)
    output = ret.stdout.readline()
    version = output.decode("utf-8").strip()
    return version


def wait_for_pod_shutdown(pod_name, namespace, timeout="10m"):
    cmd = (
        "kubectl wait --for=delete "
        f"--timeout={timeout} "
        f"-n {namespace} "
        f"pod/{pod_name}"
    )

    return run(cmd, shell=True)


def wait_for_shutdown(deployment_name, namespace, timeout="10m"):
    cmd = (
        "kubectl wait --for=delete "
        f"--timeout={timeout} "
        f"-n {namespace} "
        f"deploy/{deployment_name}"
    )

    return run(cmd, shell=True)


def get_pod_name_for_sdep(sdep_name, namespace, attempts=20, sleep=5):
    for _ in range(attempts):
        ret = run(
            f"kubectl get -n {namespace} pod -l seldon-deployment-id={sdep_name} -o json",
            shell=True,
            stdout=subprocess.PIPE,
        )
        if ret.returncode == 0:
            logging.info(f"Successfully waited for pod for {sdep_name}")
            break
        logging.warning(
            f"Unsuccessful wait command but retrying for SeldonDeployment pod {sdep_name}"
        )
        time.sleep(sleep)
    assert ret.returncode == 0, "Failed to get  pod names: non-zero return code"
    data = json.loads(ret.stdout)
    pod_names = []
    for item in data["items"]:
        pod_names.append(item["metadata"]["name"])
    logging.info(
        f"For SeldonDeployment {sdep_name} " f"found following pod: {pod_names}"
    )
    return pod_names


def log_sdep_logs(sdep_name, namespace, attempts=20, sleep=5):
    pod_names = get_pod_name_for_sdep(sdep_name, namespace, attempts, sleep)
    for pod_name in pod_names:
        for _ in range(attempts):
            ret = run(
                f"kubectl logs -n {namespace} {pod_name} --all-containers=true",
                shell=True,
                stdout=subprocess.PIPE,
            )
            if ret.returncode == 0:
                logging.info(f"Successfully got logs for {pod_name}")
                break
            logging.warning(f"Unsuccessful kubectl logs for {pod_name} but retrying")
            time.sleep(sleep)
        assert ret.returncode == 0, f"Failed to get logs for {pod_name}"
        logging.warning(ret.stdout.decode())


def get_deployment_names(sdep_name, namespace, attempts=20, sleep=5):
    for _ in range(attempts):
        ret = run(
            f"kubectl get -n {namespace} sdep {sdep_name} -o json",
            shell=True,
            stdout=subprocess.PIPE,
        )
        if ret.returncode == 0:
            logging.info(f"Successfully waited for SeldonDeployment {sdep_name}")
            break
        logging.warning(
            f"Unsuccessful wait command but retrying for SeldonDeployment {sdep_name}"
        )
        time.sleep(sleep)
    assert ret.returncode == 0, "Failed to get deployment names: non-zero return code"
    data = json.loads(ret.stdout)
    # The `deploymentStatus` is dictionary which keys are names of deployments
    deployment_names = list(data.get("status", {}).get("deploymentStatus", {}))
    logging.info(
        f"For SeldonDeployment {sdep_name} "
        f"found following deployments: {deployment_names}"
    )
    return deployment_names


def wait_for_deployment(deployment_name, namespace, attempts=50, sleep=5):
    logging.info(f"Waiting for deployment {deployment_name}")
    for _ in range(attempts):
        ret = run(
            f"kubectl rollout status -n {namespace} deploy/{deployment_name}",
            shell=True,
        )
        if ret.returncode == 0:
            logging.info(f"Successfully waited for deployment {deployment_name}")
            break
        logging.warning(f"Unsuccessful wait command but retrying for {deployment_name}")
        time.sleep(sleep)
    assert (
        ret.returncode == 0
    ), f"Wait for rollout of {deployment_name} failed: non-zero return code"


def wait_for_rollout(
    sdep_name, namespace, attempts=50, sleep=5, expected_deployments=1
):
    deployment_names = []
    for _ in range(attempts):
        deployment_names = get_deployment_names(sdep_name, namespace)
        deployments = len(deployment_names)

        if deployments == expected_deployments:
            break
        time.sleep(sleep)

    error_msg = (
        f"Expected {expected_deployments} deployment(s) but got {len(deployment_names)}"
    )
    assert len(deployment_names) == expected_deployments, error_msg

    for deployment_name in deployment_names:
        wait_for_deployment(deployment_name, namespace, attempts, sleep)


def retry_run(cmd, attempts=10, sleep=5):
    for i in range(attempts):
        ret = run(cmd, shell=True)
        if ret.returncode == 0:
            logging.info(f"Successfully ran command: {cmd}")
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
        # should prob be checking for Failed but https://github.com/SeldonIO/seldon-core/issues/2044
        if ("status" in data) and (data["status"]["state"] == "Available"):
            logging.info(f"Status for SeldonDeployment {name} is ready.")
            return data
        else:
            logging.warning("Failed to find status - sleeping")
            time.sleep(sleep)


def get_pod_names(deployment_name, namespace):
    cmd = f"kubectl get pod -l app={deployment_name} -n {namespace} -o json"
    ret = run(cmd, shell=True, check=True, stdout=subprocess.PIPE)
    pods = json.loads(ret.stdout)

    pod_names = []
    for pod in pods["items"]:
        pod_metadata = pod["metadata"]
        pod_name = pod_metadata["name"]
        pod_names.append(pod_name)

    return pod_names


def rest_request(
    model,
    namespace,
    endpoint=API_AMBASSADOR,
    data_size=5,
    rows=1,
    data=None,
    dtype="tensor",
    names=None,
    method="predict",
    predictor_name="default",
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
            method=method,
            predictor_name=predictor_name,
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
    method="predict",
    predictor_name="default",
):
    sleeping_times = [1, 5, 10]
    attempt = 0
    finished = False
    r = None
    while not finished:
        r = rest_request(
            model,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
            dtype=dtype,
            names=names,
            method=method,
            predictor_name=predictor_name,
        )

        if r is None or r.status_code != 200:
            if attempt >= len(sleeping_times):
                finished = True
            else:
                sleep = sleeping_times[attempt]
                logging.info(f"Sleeping {sleep} sec and trying again")
                time.sleep(sleep)
                attempt += 1
        else:
            finished = True

    return r


def initial_grpc_request(
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
        return grpc_request_ambassador(
            model,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
        )
    except Exception:
        logging.warning("Sleeping 1 sec and trying again")
        time.sleep(1)
        try:
            return grpc_request_ambassador(
                model,
                namespace,
                endpoint=endpoint,
                data_size=data_size,
                rows=rows,
                data=data,
            )
        except Exception:
            logging.warning("Sleeping 5 sec and trying again")
            time.sleep(5)
            try:
                return grpc_request_ambassador(
                    model,
                    namespace,
                    endpoint=endpoint,
                    data_size=data_size,
                    rows=rows,
                    data=data,
                )
            except Exception:
                logging.warning("Sleeping 10 sec and trying again")
                time.sleep(10)
                return grpc_request_ambassador(
                    model,
                    namespace,
                    endpoint=endpoint,
                    data_size=data_size,
                    rows=rows,
                    data=data,
                )


def create_random_data(data_size, rows=1):
    shape = [rows, data_size]
    arr = np.random.rand(rows * data_size)
    return (shape, arr)


@retry(wait=wait_exponential(max=10), stop=stop_after_attempt(5))
def rest_request_ambassador(
    deployment_name,
    namespace,
    endpoint=API_AMBASSADOR,
    data_size=5,
    rows=1,
    data=None,
    dtype="tensor",
    names=None,
    method="predict",
    predictor_name="default",
    model_name="classifier",
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
    elif dtype == "strData":
        payload = {"strData": arr}
    else:
        payload = {"data": {"ndarray": arr}}

    if names is not None:
        payload["data"]["names"] = names

    if method == "predict":
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
    elif method == "explain":
        response = requests.post(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + "-explainer"
            + "/"
            + predictor_name
            + "/api/v0.1/explain",
            json=payload,
        )
    elif method == "metadata":
        response = requests.get(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + "/api/v0.1/metadata/"
            + model_name
        )
    elif method == "graph-metadata":
        response = requests.get(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + "/api/v1.0/metadata"
        )
    elif method == "openapi_ui":
        response = requests.get(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + "/api/v0.1/doc/"
        )
    elif method == "openapi_schema":
        response = requests.get(
            "http://"
            + endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + "/api/v0.1/doc/seldon.json"
        )

    return response


@retry(wait=wait_exponential(max=10), stop=stop_after_attempt(5))
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


def grpc_request_ambassador(
    deployment_name, namespace, endpoint=API_AMBASSADOR, data_size=5, rows=1, data=None
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
    try:
        response = stub.Predict(request=request, metadata=metadata)
        channel.close()
        return response
    except Exception as e:
        channel.close()
        raise e


def grpc_request_ambassador_metadata(
    deployment_name, namespace, endpoint=API_AMBASSADOR, model_name=None
):
    if model_name is None:
        request = empty_pb2.Empty()
    else:
        request = prediction_pb2.SeldonModelMetadataRequest(name=model_name)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [("seldon", deployment_name)]
    else:
        metadata = [("seldon", deployment_name), ("namespace", namespace)]
    try:
        if model_name is None:
            response = stub.GraphMetadata(request=request, metadata=metadata)
        else:
            response = stub.ModelMetadata(request=request, metadata=metadata)
        channel.close()
        return response
    except Exception as e:
        channel.close()
        raise e


def grpc_request_ambassador2(
    deployment_name,
    namespace,
    endpoint="localhost:8004",
    data_size=5,
    rows=1,
    data=None,
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
    except Exception:
        logging.warning("Warning - caught exception")
        return grpc_request_ambassador(
            deployment_name,
            namespace,
            endpoint=endpoint,
            data_size=data_size,
            rows=rows,
            data=data,
        )


def clean_string(string):
    string = string.lower()
    string = re.sub(r"\]$", "", string)
    string = re.sub(r"[_\[\./]", "-", string)
    return string


def assert_model_during_op(op, *assert_args, **assert_kwargs):
    with ThreadPoolExecutor(max_workers=1) as executor:
        future = executor.submit(op)

        try:
            while future.running():
                assert_model(*assert_args, **assert_kwargs)
        except AssertionError as err:
            # In case the assertion failed, try to cancel Future or wait for it
            # to finish before raising
            cancelled = future.cancel()
            if not cancelled:
                wait([future])

            raise err

        # Future.exception() will raise any exceptions thrown within the future
        future.exception()


def assert_model(sdep_name, namespace, initial=False, endpoint=API_AMBASSADOR):
    _request = initial_rest_request if initial else rest_request
    r = _request(sdep_name, namespace, endpoint=endpoint)

    assert r is not None
    assert r.status_code == 200

    # Assert possible return values across different models
    response = r.json()
    values = response["data"]["tensor"]["values"]
    assert values in [
        [1.0, 2.0, 3.0, 4.0],  # fixed-model:0.1
        [5.0, 6.0, 7.0, 8.0],  # fixed-model:0.2
    ]

    # NOTE: The following will test if the `SeldonDeployment` can be fetched as
    # a Kubernetes resource. This covers cases where some resources (e.g. CRD
    # versions or webhooks) may get inadvertently removed between versions.
    # not checking status here as wait_for_status called previously
    ret = run(
        f"kubectl get -n {namespace} sdep {sdep_name}",
        stdout=subprocess.DEVNULL,
        shell=True,
    )
    assert ret.returncode == 0


def to_resources_path(file_name):
    return os.path.join(RESOURCES_PATH, file_name)


def post_comment_in_pr(body, check=False):
    try:
        run(f'jx gitops pr comment --comment "{body}"', shell=True, check=True)
    except Exception:
        logging.exception("Error posting comment with results")
        if check:
            raise


def create_and_run_script(folder, notebook):
    run(
        f"jupyter nbconvert --template ../../notebooks/convert.tpl --to script {folder}/{notebook}.ipynb",
        shell=True,
        check=True,
    )
    run(f"chmod u+x {folder}/{notebook}.py", shell=True, check=True)
    try:
        run(
            f"cd {folder} && ./{notebook}.py",
            shell=True,
            check=True,
            encoding="utf-8",
        )
    except CalledProcessError as e:
        logging.error(
            f"failed notebook test {notebook} stdout:{e.stdout}, stderr:{e.stderr}"
        )
        run("kubectl delete sdep --all", shell=True, check=False)
        raise e


def bench_results_from_output_logs(name, namespace="argo", print_results=True):

    output = run(f"argo logs --no-color {name} -n {namespace}",
                 capture_output=True, encoding="utf-8", check=True, shell=True)

    output.check_returncode()

    log_array = output.stdout.split("\n")

    results = []
    for log in log_array:
        if "latenc" in log:
            # Only process if contains results of benchmark
            log_clean = json.loads(":".join(log.split(":")[1:]))
            result = parse_bench_results_from_log(log_clean, print_results=print_results)
            results.append(result)

    return results


def parse_bench_results_from_log(results_log, print_results=True,):
    final = {}
    # For GHZ / grpc
    if "average" in results_log:
        final["mean"] = results_log["average"] / 1e6
        if results_log.get("latencyDistribution", False):
            final["50th"] = results_log["latencyDistribution"][-5]["latency"] / 1e6
            final["90th"] = results_log["latencyDistribution"][-3]["latency"] / 1e6
            final["95th"] = results_log["latencyDistribution"][-2]["latency"] / 1e6
            final["99th"] = results_log["latencyDistribution"][-1]["latency"] / 1e6
        final["throughputAchieved"] = results_log["rps"]
        final["success"] = results_log["statusCodeDistribution"].get("OK", 0)
        final["errors"] = sum(results_log["statusCodeDistribution"].values()) - final["success"]
    # For vegeta / rest
    else:
        final["mean"] = results_log["latencies"]["mean"] / 1e6
        final["50th"] = results_log["latencies"]["50th"] / 1e6
        final["90th"] = results_log["latencies"]["90th"] / 1e6
        final["95th"] = results_log["latencies"]["95th"] / 1e6
        final["99th"] = results_log["latencies"]["99th"] / 1e6
        final["throughputAchieved"] = results_log["throughput"]
        final["success"] = results_log["status_codes"].get("200", 0)
        final["errors"] = sum(results_log["status_codes"].values()) - final["success"]
    for k in results_log["params"].keys():
        final[k] = results_log["params"][k]
    if print_results:
        logging.warning("-----")
        logging.warning("ParamNames:", results_log["params"].keys())
        logging.warning("ParamNames:", results_log["params"].values())
        logging.warning("\tLatencies:")
        logging.warning("\t\tmean:", final["mean"], "ms")
        logging.warning("\t\t50th:", final["50th"], "ms")
        logging.warning("\t\t90th:", final["90th"], "ms")
        logging.warning("\t\t95th:", final["95th"], "ms")
        logging.warning("\t\t99th:", final["99th"], "ms")
        logging.warning("")
        logging.warning("\tRate:", str(final["throughputAchieved"]) + "/s")
        logging.warning("\tSuccess:", final["success"])
        logging.warning("\tErrors:", final["errors"])
    return final


def run_benchmark_and_capture_results(
            name="seldon-batch-job",
            namespace="argo",
            parallelism=BENCHMARK_PARALLELISM,
            replicas_list=["1"],
            server_workers_list=["5"],
            server_threads_list=["1"],
            model_uri_list=["gs://seldon-models/sklearn/iris"],
            server_list=["SKLEARN_SERVER"],
            api_type_list=["rest"],
            requests_cpu_list=["2000Mi"],
            requests_memory_list=["500Mi"],
            limits_cpu_list=["2000Mi"],
            limits_memory_list=["500Mi"],
            disable_orchestrator_list=["false"],
            benchmark_cpu_list=["1"],
            benchmark_concurrency_list=["1"],
            benchmark_duration_list=["30s"],
            benchmark_rate_list=["0"],
            benchmark_data={"data": {"ndarray": [[1, 2, 3, 4]]}},
        ):

    data_str = (json.dumps(benchmark_data)
                    .replace("{", "\\{")
                    .replace("}", "\\}")
                    .replace(",", "\\,"))

    delim = "|"

    replicas = delim.join(replicas_list)
    server_workers = delim.join(server_workers_list)
    server_threads = delim.join(server_threads_list)
    model_uri = delim.join(model_uri_list)
    server = delim.join(server_list)
    api_type = delim.join(api_type_list)
    requests_cpu = delim.join(requests_cpu_list)
    requests_memory = delim.join(requests_memory_list)
    limits_cpu = delim.join(limits_cpu_list)
    limits_memory = delim.join(limits_memory_list)
    disable_orchestrator = delim.join(disable_orchestrator_list)
    benchmark_cpu = delim.join(benchmark_cpu_list)
    benchmark_concurrency = delim.join(benchmark_concurrency_list)
    benchmark_duration = delim.join(benchmark_duration_list)
    benchmark_rate = delim.join(benchmark_rate_list)

    kwargs = {
        "shell": True,
        "check": True,
    }

    run(f'''
        helm template seldon-benchmark-workflow ../../helm-charts/seldon-benchmark-workflow/ \\
            --set workflow.namespace="{namespace}" \\
            --set workflow.name="{name}" \\
            --set workflow.parallelism="{parallelism}" \\
            --set seldonDeployment.name="{name}-sdep" \\
            --set seldonDeployment.replicas="{replicas}" \\
            --set seldonDeployment.serverWorkers="{server_workers}" \\
            --set seldonDeployment.serverThreads="{server_threads}" \\
            --set seldonDeployment.modelUri="{model_uri}" \\
            --set seldonDeployment.server="{server}" \\
            --set seldonDeployment.apiType="{api_type}" \\
            --set seldonDeployment.requests.cpu="{requests_cpu}" \\
            --set seldonDeployment.requests.memory="{requests_memory}" \\
            --set seldonDeployment.limits.cpu="{limits_cpu}" \\
            --set seldonDeployment.limits.memory="{limits_memory}" \\
            --set seldonDeployment.disableOrchestrator="{disable_orchestrator}" \\
            --set benchmark.cpu="{benchmark_cpu}" \\
            --set benchmark.concurrency="{benchmark_concurrency}" \\
            --set benchmark.duration="{benchmark_duration}" \\
            --set benchmark.rate="{benchmark_rate}" \\
            --set benchmark.data='{data_str}' \\
            | argo submit -
        ''',
        **kwargs)
    run("argo list -n {namespace}", **kwargs)
    run(f"argo logs -n {namespace} -f {name}", **kwargs)

    results = bench_results_from_output_logs(name)

    df_results = pd.DataFrame.from_dict(results)

    run(f"argo delete -n {namespace} {name}", **kwargs)

    return df_results

