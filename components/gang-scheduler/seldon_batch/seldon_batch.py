from threading import Thread, Lock
from queue import Queue
import requests
import os
import json
from seldon_core.seldon_client import SeldonClient


def start_batch_processing_loop(
    deployment_name,
    gateway_type,
    namespace,
    endpoint,
    transport,
    payload_type,
    parallelism,
    retries,
    input_data_path,
    output_data_path,
    method,
    log_level,
):

    url = f"http://{endpoint}/seldon/{namespace}/{deployment_name}/api/v1.0/{method}"
    q_in = Queue(parallelism * 2)
    q_out = Queue(parallelism * 2)

    def _start_request_worker():
        with requests.Session() as session:
            while True:
                line = q_in.get()
                headers = {"Content-Type": "application/json"}
                # TODO: Use Seldon Client instead after optimising it
                response = session.post(url, body=line, headers=headers)

                q_out.put(response.text)
                q_in.task_done()

    def _start_file_worker():
        output_data_file = open(output_data_path, "w")
        while True:
            line = q_out.get()
            output_data_file.write(f"{line}")
            q_out.task_done()

    for _ in range(parallelism):
        t = Thread(target=_start_request_worker)
        t.daemon = True
        t.start()

    t = Thread(target=_start_file_worker)
    t.daemon = True
    t.start()

    input_data_file = open(input_data_path, "r")
    for line in input_data_file:
        q_in.put(line)

    q_in.join()
    q_out.join()
