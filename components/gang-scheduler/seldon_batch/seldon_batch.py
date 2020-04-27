import json
import numpy as np
from seldon_core.seldon_client import SeldonClient


# def start_batch_processing_loop(
#    deployment_name,
#    gateway_type,
#    namespace,
#    endpoint,
#    transport,
#    payload_type,
#    parallelism,
#    retries,
#    input_data_path,
#    output_data_path,
#    method,
#    log_level,
# ):
#    sc = SeldonClient(
#        deployment_name=deployment_name,
#        gateway=gateway_type,
#        namespace=namespace,
#        gateway_endpoint=endpoint,
#        transport=transport,
#        payload_type="ndarray",
#    )
#    input_data_file = open(input_data_path, "r")
#    output_data_file = open(output_data_path, "w")
#    # TODO: introduce parallelim with Queue
#    for line in input_data_file:
#        raw_data = json.loads(line)
#        data = np.array(raw_data)
#        output = sc.predict(data=data)
#        # TODO: Add identifier to track back to input
#        # TODO: HAndler errors
#        output_data_file.write(f"{output.response}\n")

from threading import Thread, Lock
from queue import Queue
import requests
import os


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
    # lock = Lock()
    # file_cursor = os.path.getsize(input_file_path)

    def _start_request_worker():
        while True:
            line = q_in.get()
            data = json.loads(line)
            response = requests.post(url, json=data)

            q_out.put(response.text)
            q_in.task_done()
            ## We add a lock to ensure there is no race condition on setting end of file cursor
            # with lock:
            #    # Ensure cursor moves across the file based on length of line
            #    file_cursor -= len(line)
            #    q_out.put((response.text, bool(file_cursor)))

    def _start_file_worker():
        output_data_file = open(output_data_path, "w")
        while True:
            line = q_out.get()
            output_data_file.write(f"{line}\n")
            q_out.task_done()
            ## If we've reached the last line then we break
            # if not file_cursor:
            #    break

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
