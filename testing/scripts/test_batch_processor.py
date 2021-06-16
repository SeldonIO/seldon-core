import json
import logging
import time
import uuid
from subprocess import run

import requests

from seldon_core.batch_processor import start_multithreaded_batch_worker
from seldon_e2e_utils import (
    API_ISTIO_GATEWAY,
    create_random_data,
    initial_rest_request,
    rest_request,
    rest_request_ambassador,
    retry_run,
    wait_for_rollout,
    wait_for_status,
)


class TestBatchWorker(object):
    def test_batch_worker(self, namespace):
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(10)

        batch_size = 1000
        input_data_path = "input-data.txt"
        output_data_path = "output-data.txt"

        with open(input_data_path, "w") as f:
            for i in range(batch_size):
                f.write("[[1,2,3,4]]\n")

        start_multithreaded_batch_worker(
            "sklearn",
            "istio",
            namespace,
            API_ISTIO_GATEWAY,
            "rest",
            "data",
            "ndarray",
            100,
            3,
            1,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
        )

        with open(output_data_path, "r") as f:
            for line in f:
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("ndarray", False)

        # Now test that a batch size of 10 works
        start_multithreaded_batch_worker(
            "sklearn",
            "istio",
            namespace,
            API_ISTIO_GATEWAY,
            "rest",
            "data",
            "ndarray",
            100,
            3,
            10,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
        )

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count +=1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("ndarray", False)
            assert (count == batch_size)

        logging.info("Success for test_batch_worker")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
