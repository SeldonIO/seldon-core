import json
import numpy as np
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
    sc = SeldonClient(
        deployment_name=deployment_name,
        gateway=gateway_type,
        namespace=namespace,
        gateway_endpoint=endpoint,
        transport=transport,
        payload_type="ndarray",
    )
    input_data_file = open(input_data_path, "r")
    output_data_file = open(output_data_path, "w")
    # TODO: introduce parallelim with Queue
    for line in input_data_file:
        raw_data = json.loads(line)
        data = np.array(raw_data)
        output = sc.predict(data=data)
        # TODO: Add identifier to track back to input
        # TODO: HAndler errors
        output_data_file.write(f"{output.response}\n")
