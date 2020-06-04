import click
import json
import requests
from queue import Queue
from threading import Thread
from seldon_core.storage import Storage
import os
import uuid

CHOICES_GATEWAY_TYPE = ["ambassador", "istio", "seldon"]
CHOICES_TRANSPORT = ["rest", "grpc"]
CHOICES_PAYLOAD_TYPE = ["data", "json", "bytes", "str"]
CHOICES_DATA_TYPE = ["ndarray", "tensor", "tftensor"]
CHOICES_METHOD = ["predictions", "explain"]
CHOICES_LOG_LEVEL = ["debug", "info", "warning", "error"]

# Create uuid file
DATA_TEMP_DIRPATH = os.path.join(__file__, "TEMP_DATA_FILES")
DATA_TEMP_UUID = str(uuid.uuid4())
DATA_TEMP_INPUT_FILENAME = f"{DATA_TEMP_UUID}-input.txt"
DATA_TEMP_OUTPUT_FILENAME = f"{DATA_TEMP_UUID}-output.txt"


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
    default="predictions",
)
@click.option(
    "--log-level",
    "-l",
    envvar="SELDON_BATCH_LOG_LEVEL",
    type=click.Choice(CHOICES_LOG_LEVEL),
    default="info",
)
@click.option(
    "--generate-id",
    "-u",
    envvar="SELDON_BATCH_LOG_LEVEL",
    type=click.Choice(CHOICES_LOG_LEVEL),
    default="info",
)
def run_cli(
    deployment_name,
    gateway_type,
    namespace,
    host,
    transport,
    payload_type,
    workers,
    retries,
    input_data_path,
    output_data_path,
    method,
    log_level,
):
    is_remote_input_path = Storage.is_remote_path(input_data_path)
    is_remote_output_path = Storage.is_remote_path(output_data_path)

    local_output_data_path = None
    if is_remote_output_path:
        try:
            os.mkdir(DATA_TEMP_DIRPATH)
        except FileExistsError:
            pass
        local_output_data_path = os.path.join(
            DATA_TEMP_DIRPATH, DATA_TEMP_OUTPUT_FILENAME
        )
    else:
        # Check if file path is correct and is not directory by creating temp file
        with open(output_data_path, "x") as tempfile:
            pass
        local_output_data_path = output_data_path

    local_input_data_path = None
    if is_remote_input_path:
        local_input_data_path = os.path.join(
            DATA_TEMP_DIRPATH, DATA_TEMP_INPUT_FILENAME
        )
        Storage.download(input_data_path, DATA_TEMP_INPUT_FILENAME)
        if os.path.isdir(DATA_TEMP_INPUT_FILENAME):
            raise RuntimeError(
                "Only single files are supported - "
                f"directory {input_data_path} is not valid. "
                "Please provide a file, not a directory."
            )
    else:
        local_input_data_path = input_data_path

    # TODO: Add checks that url is valid
    url = f"http://{host}/seldon/{namespace}/{deployment_name}/api/v1.0/{method}"
    q_in = Queue(workers * 2)
    q_out = Queue(workers * 2)

    def _start_request_worker():
        with requests.Session() as session:
            while True:
                line = q_in.get()
                headers = {"Content-Type": "application/json"}
                # TODO: Use Seldon Client instead after optimising it
                # TODO: Deal with failed requests
                response = session.post(url, data=line, headers=headers)

                q_out.put(response.text)
                q_in.task_done()

    def _start_file_worker():
        output_data_file = open(local_output_data_path, "w")
        while True:
            line = q_out.get()
            output_data_file.write(f"{line}")
            q_out.task_done()

    for _ in range(workers):
        t = Thread(target=_start_request_worker)
        t.daemon = True
        t.start()

    t = Thread(target=_start_file_worker)
    t.daemon = True
    t.start()

    input_data_file = open(local_input_data_path, "r")
    for line in input_data_file:
        q_in.put(line)

    q_in.join()
    q_out.join()

    if is_remote_output_path:
        Storage.upload(local_output_data_path, output_data_path)
