import click
import json
import requests
from queue import Queue
from threading import Thread

CHOICES_GATEWAY_TYPE = ["ambassador", "istio", "seldon"]
CHOICES_TRANSPORT = ["rest", "grpc"]
CHOICES_PAYLOAD_TYPE = ["ndarray", "tensor", "tftensor", "json", "bytes", "str", "raw"]
CHOICES_METHOD = ["predictions", "explain"]
CHOICES_LOG_LEVEL = ["debug", "info", "warning", "error"]


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
    "--payload-type",
    "-p",
    envvar="SELDON_BATCH_PAYLOAD_TYPE",
    type=click.Choice(CHOICES_PAYLOAD_TYPE),
    default="raw",
)
@click.option(
    "--parallelism", "-x", envvar="SELDON_BATCH_PARALLELISM", type=int, default=1
)
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
def run_cli(
    deployment_name,
    gateway_type,
    namespace,
    host,
    transport,
    payload_type,
    parallelism,
    retries,
    input_data_path,
    output_data_path,
    method,
    log_level,
):
    url = f"http://{host}/seldon/{namespace}/{deployment_name}/api/v1.0/{method}"
    q_in = Queue(parallelism * 2)
    q_out = Queue(parallelism * 2)

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
