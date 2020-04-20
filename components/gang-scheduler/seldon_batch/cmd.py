import click
from .seldon_batch import start_batch_processing_loop

CHOICES_GATEWAY_TYPE = ["ambassador", "istio", "seldon"]
CHOICES_TRANSPORT = ["rest", "grpc"]
CHOICES_PAYLOAD_TYPE = ["ndarray", "tensor", "tftensor", "json", "bytes", "str"]
CHOICES_METHOD = ["predict", "explain"]
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
    "--endpoint",
    "-e",
    envvar="SELDON_BATCH_ENDPOINT",
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
    default="ndarray",
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
    "-i",
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
def run_cli(
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
    start_batch_processing_loop(
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
    )
