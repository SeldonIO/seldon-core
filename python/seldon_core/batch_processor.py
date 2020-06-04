import click
import json
import requests
from queue import Queue
from threading import Thread
from seldon_core.seldon_client import SeldonClient
import numpy as np
import os
import uuid

CHOICES_GATEWAY_TYPE = ["ambassador", "istio", "seldon"]
CHOICES_TRANSPORT = ["rest", "grpc"]
CHOICES_PAYLOAD_TYPE = ["data", "json", "str"]
CHOICES_DATA_TYPE = ["ndarray", "tensor", "tftensor"]
# TODO: Add explainer support
CHOICES_METHOD = ["predict"]
CHOICES_LOG_LEVEL = ["debug", "info", "warning", "error"]
CHOICES_INDEX_TYPE = ["enumerate", "unique"]


@click.command()
@click.option(
    "--deployment-name", "-d", envvar="SELDON_BATCH_DEPLOYMENT_NAME", required=True
)
@click.option(
    "--gateway-type",
    "-g",
    envvar="SELDON_BATCH_GATEWAY_TYPE",
    type=click.Choice(CHOICES_GATEWAY_TYPE),
    default="istio",
)
@click.option("--namespace", "-n", envvar="SELDON_BATCH_NAMESPACE", default="default")
@click.option(
    "--host",
    "-h",
    envvar="SELDON_BATCH_HOST",
    default="istio-ingressgateway.istio-system.svc.cluster.local:80",
)
@click.option(
    "--transport",
    "-t",
    envvar="SELDON_BATCH_TRANSPORT",
    type=click.Choice(CHOICES_TRANSPORT),
    default="rest",
)
@click.option(
    "--data-type",
    "-a",
    envvar="SELDON_BATCH_DATA_TYPE",
    type=click.Choice(CHOICES_DATA_TYPE),
    default="data",
)
@click.option(
    "--payload-type",
    "-p",
    envvar="SELDON_BATCH_PAYLOAD_TYPE",
    type=click.Choice(CHOICES_PAYLOAD_TYPE),
    default="ndarray",
)
@click.option("--workers", "-w", envvar="SELDON_BATCH_WORKERS", type=int, default=1)
@click.option("--retries", "-r", envvar="SELDON_BATCH_RETRIES", type=int, default=3)
@click.option(
    "--input-data-path",
    "-i",
    envvar="SELDON_BATCH_INPUT_DATA_PATH",
    type=click.Path(),
    default="/assets/input-data.txt",
)
@click.option(
    "--output-data-path",
    "-o",
    envvar="SELDON_BATCH_OUTPUT_DATA_PATH",
    type=click.Path(),
    default="/assets/input-data.txt",
)
@click.option(
    "--method",
    "-m",
    envvar="SELDON_BATCH_METHOD",
    type=click.Choice(CHOICES_METHOD),
    default="predict",
)
@click.option(
    "--log-level",
    "-l",
    envvar="SELDON_BATCH_LOG_LEVEL",
    type=click.Choice(CHOICES_LOG_LEVEL),
    default="info",
)
@click.option(
    "--index-type",
    "-x",
    envvar="SELDON_BATCH_INDEX_TYPE",
    type=click.Choice(CHOICES_INDEX_TYPE),
    default="info",
)
def run_cli(
    deployment_name,
    gateway_type,
    namespace,
    host,
    transport,
    data_type,
    payload_type,
    workers,
    retries,
    input_data_path,
    output_data_path,
    method,
    log_level,
    index_type,
):
    q_in = Queue(workers * 2)
    q_out = Queue(workers * 2)

    sc = SeldonClient(
        gateway=gateway_type,
        transport=transport,
        deployment_name=deployment_name,
        payload_type=payload_type,
        gateway_endpoint=host,
        namespace=namespace,
        client_return_type="dict",
    )

    def _start_request_worker():
        while True:
            batch_uid, batch_idx, input_raw = q_in.get()
            data = json.loads(input_raw)

            predict_kwargs = {}
            meta = {"tags": {"batch_uid": batch_uid, "batch_idx": batch_idx}}
            predict_kwargs["meta"] = meta
            predict_kwargs["headers"] = {"Seldon-Puid": batch_uid}

            # TODO: Add functionality to send "raw" data
            if payload_type == "data":
                # TODO: Update client to avoid requiring a numpy array
                data_np = np.array(data)
                predict_kwargs["data"] = data_np
            elif payload_type == "str":
                predict_kwargs["str_data"] = data
            elif payload_type == "json":
                predict_kwargs["json_data"] = data

            str_output = None
            for _ in range(retries):
                try:
                    # TODO: Add functionality for explainer
                    #   as explainer currently doesn't support meta
                    # if method == "predict":
                    seldon_payload = sc.predict(**predict_kwargs)
                    assert seldon_payload.success
                    str_output = json.dumps(seldon_payload.response)
                    break
                # catch all exceptions to ensure the task is marked as done
                except Exception as e:
                    # TODO: Change to log
                    print(str(e))
                    error_resp = {
                        "status": {"info": "FAILURE", "reason": str(e), "status": 1},
                        "meta": meta,
                    }
                    str_output = json.dumps(error_resp)

            # Mark task as done in the queue to add space for new tasks
            q_out.put(str_output)
            q_in.task_done()

    def _start_file_worker():
        output_data_file = open(output_data_path, "w")
        while True:
            line = q_out.get()
            output_data_file.write(f"{line}\n")
            q_out.task_done()

    for _ in range(workers):
        t = Thread(target=_start_request_worker)
        t.daemon = True
        t.start()

    t = Thread(target=_start_file_worker)
    t.daemon = True
    t.start()

    input_data_file = open(input_data_path, "r")

    enum_idx = 0
    for line in input_data_file:
        unique_id = str(uuid.uuid1())
        q_in.put((enum_idx, unique_id, line))
        enum_idx += 1

    q_in.join()
    q_out.join()
