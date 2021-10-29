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

logging.basicConfig(level=logging.DEBUG)


class TestBatchWorker(object):
    def test_batch_worker(self, namespace):
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(10)

        batch_size = 1000
        input_data_path = "batch-standard-input-data.txt"
        output_data_path = "batch-standard-output-data.txt"

        with open(input_data_path, "w") as f:
            for i in range(batch_size):
                f.write("[[1,2,3,4]]\n")

        logging.info("Sending first batch: mini-batch size=1")

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
            0,
        )

        logging.info("Finished first batch. Checking.")

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count += 1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("ndarray", False)
            assert count == batch_size

        logging.info("Sending first batch: mini-batch size=30")

        # Now test that with a mini batch size of 30 works
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
            30,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
            0,
        )

        logging.info("Finished first batch. Checking.")

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count += 1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("ndarray", False)
            assert count == batch_size

        logging.info("Success for test_batch_worker")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    def test_batch_worker_raw_predict_ndarray(self, namespace):
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(10)

        batch_size = 1000
        input_data_path = "batch-raw-ndarray-input-data.txt"
        output_data_path = "batch-raw-ndarray-output-data.txt"

        with open(input_data_path, "w") as f:
            for i in range(batch_size):
                j = {
                    "data": {"names": ["a", "b", "c"], "ndarray": [[1, 2, 3, 4]]},
                    "meta": {"tags": {"customer-id": i}},
                }
                f.write(json.dumps(j) + "\n")

        logging.info("Sending first batch: mini-batch size=1")

        start_multithreaded_batch_worker(
            "sklearn",
            "istio",
            namespace,
            API_ISTIO_GATEWAY,
            "rest",
            "raw",
            None,
            100,
            3,
            1,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
            0,
        )

        logging.info("Finished first batch. Checking.")

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count += 1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("ndarray", False)

                # Following assert checks that customer-id custom tag from raw input has been propagated
                assert (
                    output["meta"]["tags"]["customer-id"]
                    == output["meta"]["tags"]["batch_index"]
                )

            assert count == batch_size

        logging.info("Sending first batch: mini-batch size=30")

        # Now test that with a mini batch size of 30 works
        start_multithreaded_batch_worker(
            "sklearn",
            "istio",
            namespace,
            API_ISTIO_GATEWAY,
            "rest",
            "raw",
            None,
            100,
            3,
            30,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
            0,
        )

        logging.info("Finished first batch. Checking.")

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count += 1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("ndarray", False)

                # Following assert checks that customer-id custom tag from raw input has been propagated
                assert (
                    output["meta"]["tags"]["customer-id"]
                    == output["meta"]["tags"]["batch_index"]
                )

            assert count == batch_size

        logging.info("Success for test_batch_worker")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    def test_batch_worker_raw_predict_tensor(self, namespace):
        spec = "../../servers/sklearnserver/samples/iris.yaml"
        retry_run(f"kubectl apply -f {spec} -n {namespace}")
        wait_for_status("sklearn", namespace)
        wait_for_rollout("sklearn", namespace)
        time.sleep(10)

        batch_size = 1000
        input_data_path = "batch-raw-tensor-input-data.txt"
        output_data_path = "batch-raw-tensor-output-data.txt"

        with open(input_data_path, "w") as f:
            for i in range(batch_size):
                j = {
                    "data": {
                        "names": ["a", "b", "c"],
                        "tensor": {"shape": [1, 4], "values": [1, 2, 3, 4]},
                    },
                    "meta": {"tags": {"customer-id": i}},
                }
                f.write(json.dumps(j) + "\n")

        logging.info("Sending first batch: mini-batch size=1")

        start_multithreaded_batch_worker(
            "sklearn",
            "istio",
            namespace,
            API_ISTIO_GATEWAY,
            "rest",
            "raw",
            None,
            100,
            3,
            1,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
            0,
        )

        logging.info("Finished first batch. Checking.")

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count += 1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("tensor", False)

                # Following assert checks that customer-id custom tag from raw input has been propagated
                assert (
                    output["meta"]["tags"]["customer-id"]
                    == output["meta"]["tags"]["batch_index"]
                )

            assert count == batch_size

        logging.info("Sending first batch: mini-batch size=30")

        # Now test that with a mini batch size of 30 works
        start_multithreaded_batch_worker(
            "sklearn",
            "istio",
            namespace,
            API_ISTIO_GATEWAY,
            "rest",
            "raw",
            None,
            100,
            3,
            30,
            input_data_path,
            output_data_path,
            "predict",
            "debug",
            True,
            str(uuid.uuid1()),
            0,
        )

        logging.info("Finished first batch. Checking.")

        with open(output_data_path, "r") as f:
            count = 0
            for line in f:
                count += 1
                output = json.loads(line)
                # Ensure all requests are successful
                assert output.get("data", {}).get("tensor", False)

                # Following assert checks that customer-id custom tag from raw input has been propagated
                assert (
                    output["meta"]["tags"]["customer-id"]
                    == output["meta"]["tags"]["batch_index"]
                )

            assert count == batch_size

        logging.info("Success for test_batch_worker")
        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
