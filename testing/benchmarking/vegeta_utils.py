import json
import subprocess
import time
from subprocess import Popen, run

import yaml


def run_model(model_name):
    with open(model_name, "r") as stream:
        resource = yaml.safe_load(stream)
        metaName = resource["metadata"]["name"]
        run(f"kubectl apply -f {model_name}", shell=True, check=True)
        run(
            f"kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id={metaName} -o jsonpath='{{.items[0].metadata.name}}')",
            shell=True,
        )
        for i in range(60):
            ret = Popen(
                f"kubectl get sdep {metaName} -o jsonpath='{{.status.state}}'",
                shell=True,
                stdout=subprocess.PIPE,
            )
            state = ret.stdout.readline().decode("utf-8").strip()
            if state == "Available":
                break
            time.sleep(1)
        for i in range(60):
            ret = Popen(
                f"kubectl get pods -l seldon-deployment-id={metaName} -o json",
                shell=True,
                stdout=subprocess.PIPE,
            )
            raw = ret.stdout.read().decode("utf-8")
            results = json.loads(raw)
            numPods = len(results["items"])
            if numPods == 1:
                break
            time.sleep(1)
        print(state, "with", numPods, "pods")


def run_vegeta_test(vegeta_cfg, vegeta_job, wait_time):
    with open(vegeta_job, "r") as stream:
        resource = yaml.safe_load(stream)
        metaName = resource["metadata"]["name"]
        run(f"kubectl apply -f {vegeta_cfg}", shell=True)
        run(f"kubectl create -f {vegeta_job}", shell=True)
        run(
            f"kubectl wait --for=condition=complete --timeout={wait_time} job/tf-load-test",
            shell=True,
        )
        ret = Popen(
            f"kubectl logs $(kubectl get pods -l job-name={metaName} -o  jsonpath='{{.items[0].metadata.name}}')",
            shell=True,
            stdout=subprocess.PIPE,
        )
        raw = ret.stdout.readline().decode("utf-8")
        results = json.loads(raw)
        run(f"kubectl delete -f {vegeta_cfg}", shell=True)
        run(f"kubectl delete -f {vegeta_job}", shell=True)
        return results


def print_vegeta_results(results):
    print("Latencies:")
    print("\tmean:", results["latencies"]["mean"] / 1e6, "ms")
    print("\t50th:", results["latencies"]["50th"] / 1e6, "ms")
    print("\t90th:", results["latencies"]["90th"] / 1e6, "ms")
    print("\t95th:", results["latencies"]["95th"] / 1e6, "ms")
    print("\t99th:", results["latencies"]["99th"] / 1e6, "ms")
    print("")
    print("Throughput:", str(results["throughput"]) + "/s")
    print("Errors:", len(results["errors"]) > 0)


def run_ghz_test(payload, ghz_job, wait_time):
    with open(ghz_job, "r") as stream:
        resource = yaml.safe_load(stream)
        metaName = resource["metadata"]["name"]
        run(f"kubectl create configmap tf-ghz-cfg --from-file {payload}", shell=True)
        run(f"kubectl create -f {ghz_job}", shell=True)
        run(
            f"kubectl wait --for=condition=complete --timeout={wait_time} job/tf-load-test",
            shell=True,
        )
        ret = Popen(
            f"kubectl logs $(kubectl get pods -l job-name={metaName} -o  jsonpath='{{.items[0].metadata.name}}')",
            shell=True,
            stdout=subprocess.PIPE,
        )
        raw = ret.stdout.readline().decode("utf-8")
        results = json.loads(raw)
        run(f"kubectl delete -f tf-ghz-cfg", shell=True)
        run(f"kubectl delete -f {ghz_job}", shell=True)
        return results
