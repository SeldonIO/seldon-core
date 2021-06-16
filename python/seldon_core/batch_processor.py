import json
import logging
import time
import uuid
from queue import Empty, Queue
from threading import Event, Thread
import copy

import click
import numpy as np
import requests

from seldon_core.seldon_client import SeldonClient

CHOICES_GATEWAY_TYPE = ["ambassador", "istio", "seldon"]
CHOICES_TRANSPORT = ["rest", "grpc"]
CHOICES_PAYLOAD_TYPE = ["ndarray", "tensor", "tftensor"]
CHOICES_DATA_TYPE = ["data", "json", "str", "raw"]
CHOICES_METHOD = ["predict", "feedback"]
CHOICES_LOG_LEVEL = {
    "debug": logging.DEBUG,
    "info": logging.INFO,
    "warning": logging.WARNING,
    "error": logging.ERROR,
}

logger = logging.getLogger(__name__)


def setup_logging(log_level: str):
    LOG_FORMAT = (
        "%(asctime)s - batch_processor.py:%(lineno)s - %(levelname)s:  %(message)s"
    )
    logging.basicConfig(level=CHOICES_LOG_LEVEL[log_level], format=LOG_FORMAT)


def start_multithreaded_batch_worker(
        deployment_name: str,
        gateway_type: str,
        namespace: str,
        host: str,
        transport: str,
        data_type: str,
        payload_type: str,
        workers: int,
        retries: int,
        batch_size: int,
        input_data_path: str,
        output_data_path: str,
        method: str,
        log_level: str,
        benchmark: bool,
        batch_id: str,
) -> None:
    """
    Starts the multithreaded batch worker which consists of three worker types and
    two queues; the input_file_worker which reads a file and puts all lines in an
    input queue, which are then read by the multiple request_processor_workers (the
    number of parallel workers is specified by the workers param), which puts the output
    in the output queue and then the output_file_worker which puts all the outputs in the
    output file in a thread-safe approach.

    All parameters are defined and explained in detail in the run_cli function.
    """
    setup_logging(log_level)
    start_time = time.time()
    out_queue_empty_event = Event()

    q_in = Queue(workers * 2)
    q_out = Queue(workers * 2)

    if method == "feedback" and data_type != "raw":
        raise RuntimeError("Feedback method is supported only with `raw` data type.")
    elif data_type != "data" and batch_size > 1:
        raise RuntimeError("Batch size greater than 1 is only supported for `data` data type.")
    elif data_type == "raw" and method != "feedback":
        raise RuntimeError("Raw input is currently only support for feedback method.")

    sc = SeldonClient(
        gateway=gateway_type,
        transport=transport,
        deployment_name=deployment_name,
        payload_type=payload_type,
        gateway_endpoint=host,
        namespace=namespace,
        client_return_type="dict",
    )

    t_in = Thread(
        target=_start_input_file_worker, args=(q_in, input_data_path, batch_size), daemon=True
    )
    t_in.start()

    for _ in range(workers):
        Thread(
            target=_start_request_worker,
            args=(q_in, q_out, data_type, sc, method, retries, batch_id, payload_type),
            daemon=True,
        ).start()

    t_out = Thread(
        target=_start_output_file_worker,
        args=(q_out, output_data_path, out_queue_empty_event),
    )
    t_out.start()

    # Make sure all data was loaded
    t_in.join()

    # Make sure all data was passed through both queues
    q_in.join()
    q_out.join()

    # Set event so output worker can close file once it's done with q_out queue
    out_queue_empty_event.set()

    # Wait for output worker to join main thread
    t_out.join()

    if benchmark:
        logger.info(f"Elapsed time: {time.time() - start_time}")


def _start_input_file_worker(q_in: Queue, input_data_path: str, batch_size: int) -> None:
    """
    Runs logic for the input file worker which reads the input file from filestore
    and puts all of the lines into the input queue so it can be processed.

    Parameters
    ---
    q_in
        The queue to put all the data into for further processing
    input_data_path
        The local file to read the data from to be processed
    """
    input_data_file = open(input_data_path, "r")
    enum_idx = 0
    batch = []
    for line in input_data_file:
        unique_id = str(uuid.uuid1())
        batch.append((enum_idx, unique_id, line))
        # If the batch to send is the size then push to queue and rest batch
        if len(batch) == batch_size:
            q_in.put(batch)
            batch = []
        enum_idx += 1


def _start_output_file_worker(
        q_out: Queue, output_data_path: str, stop_event: Event
) -> None:
    """
    Runs logic for the output file worker which receives all the processed output
    from the request worker through the queue and adds it into the output file in a
    thread safe manner.

    Parameters
    ---
    q_out
        The queue to read the results from
    output_data_path
        The local file to write the results into
    """

    counter = 0
    with open(output_data_path, "w") as output_data_file:
        while not stop_event.is_set():
            try:
                line = q_out.get(timeout=0.1)
            except Empty:
                continue
            output_data_file.write(f"{line}\n")
            q_out.task_done()

            counter += 1
            if counter % 100 == 0:
                logger.info(f"Processed instances: {counter}")
    logger.info(f"Total processed instances: {counter}")


def _start_request_worker(
        q_in: Queue,
        q_out: Queue,
        data_type: str,
        sc: SeldonClient,
        method: str,
        retries: int,
        batch_id: str,
        payload_type: str,
) -> None:
    """
    Runs logic for the worker that sends requests from the queue until the queue
    gets completely empty. The worker marks the task as done when it finishes processing
    to ensure that the queue gets populated as it's currently configured with a threshold.

    Parameters
    ---
    q_in
        Queue to read the input data from
    q_out
        Queue to put the resulting requests into
    data_type
        The json/str/data type to send the requests as
    sc
        An initialised Seldon Client configured to send the requests to
    method:
        Method to call: predict or feedback
    retries
        The number of attempts to try for each request
    batch_id
        The unique identifier for the batch which is passed to all requests
    """
    while True:
        input_data = q_in.get()
        if method == "predict":
            # If we have a batch size > 1 then we wish to use the other method and add each output to queue
            if len(input_data) > 1:
                str_outputs = _send_batch_predict_multi_request(
                    input_data,
                    data_type,
                    sc,
                    retries,
                    batch_id,
                    payload_type,
                )
                for str_output in str_outputs:
                    q_out.put(str_output)
                # continue with next input
                q_in.task_done()
                continue
            batch_idx, batch_instance_id, input_raw = input_data[0]
            str_output = _send_batch_predict(
                batch_idx,
                batch_instance_id,
                input_raw,
                data_type,
                sc,
                retries,
                batch_id,
            )
        elif method == "feedback":
            str_output = _send_batch_feedback(
                input_data,
                sc,
                retries,
                batch_id,
            )
        # Mark task as done in the queue to add space for new tasks
        q_out.put(str_output)
        q_in.task_done()

def _send_batch_predict_multi_request(
        input_raw: [],
        data_type: str,
        sc: SeldonClient,
        retries: int,
        batch_id: str,
        payload_type: str,
) -> str:
    """
    Send an request using the Seldon Client with batch context including the
    unique ID of the batch and the Batch enumerated index as metadata. This
    function also uses the unique batch ID as request ID so the request can be
    traced back individually in the Seldon Request Logger context. Each request
    will be attempted for the number of retries, and will return the string
    serialised result.
    Parameters
    ---
    # TODO: Change this
    input_raw
        The raw input in string format to be loaded to the respective format
    data_type
        The data type to send which can be str, json and data
    sc
        The instance of SeldonClient to use to send the requests to the seldon model
    retries
        The number of times to retry the request
    batch_id
        The unique identifier for the batch which is passed to all requests
    Returns
    ---
        A string serialised result of the response (or equivalent data with error info)
    """

    instance_ids = [x[1] for x in input_raw]
    indexes = [x[0] for x in input_raw]

    predict_kwargs = {}
    meta = {
        "tags": {
            "batch_id": batch_id,
        }
    }

    predict_kwargs["meta"] = meta
    predict_kwargs["headers"] = {"Seldon-Puid": instance_ids[0]}

    try:
        if data_type == "data":
            data = json.loads(input_raw[0][2])
            data_np = np.array(data)
            overall = data_np
            for i, raw_data in enumerate(input_raw):
                if i == 0:
                    continue
                data = json.loads(raw_data[2])
                data_np = np.array(data)
                overall = np.concatenate((overall, data_np))
            predict_kwargs["data"] = overall
        data = json.loads(input_raw[0][2])
        if data_type == "str":
            predict_kwargs["str_data"] = data
        elif data_type == "json":
            predict_kwargs["json_data"] = data

        response = None
        for i in range(retries):
            try:
                seldon_payload = sc.predict(**predict_kwargs)
                assert seldon_payload.success
                response = seldon_payload.response
                break
            except (requests.exceptions.RequestException, AssertionError) as e:
                logger.error(f"Exception: {e}, retries {retries}")
                if i == (retries - 1):
                    raise

    except Exception as e:
        error_resp = {
            "status": {"info": "FAILURE", "reason": str(e), "status": 1},
            "meta": meta,
        }
        print(e)
        str_output = json.dumps(error_resp)
        return [str_output]

    # Take the response create new responses for each request
    responses = []
    for i in range(len(input_raw)):
        newResponse = copy.deepcopy(response)
        if payload_type == "ndarray":
            newResponse["data"]["ndarray"] = [response["data"]["ndarray"][i]]
            newResponse["meta"]["tags"]["tags"]["batch_index"] = indexes[i]
            newResponse["meta"]["tags"]["tags"]["batch_instance_id"] = instance_ids[i]
            responses.append(json.dumps(newResponse))
        elif payload_type == "tensor":
            tensor = np.array(response["data"]["tensor"]["values"])
            shape = response["data"]["tensor"]["shape"]
            ndarray = tensor.reshape(shape)
            newResponse["data"]["tensor"]["shape"][0] = 1
            newResponse["data"]["tensor"]["values"] = [np.ndarray.tolist(ndarray[i])]
            newResponse["meta"]["tags"]["tags"]["batch_index"] = indexes[i]
            newResponse["meta"]["tags"]["tags"]["batch_instance_id"] = instance_ids[i]
            responses.append(json.dumps(newResponse))
        else:
            raise RuntimeError("Only `ndarray` and `tensor` input is currently supported for batch size greater than 1.")

    return responses


def _send_batch_predict(
        batch_idx: int,
        batch_instance_id: int,
        input_raw: str,
        data_type: str,
        sc: SeldonClient,
        retries: int,
        batch_id: str,
) -> str:
    """
    Send an request using the Seldon Client with batch context including the
    unique ID of the batch and the Batch enumerated index as metadata. This
    function also uses the unique batch ID as request ID so the request can be
    traced back individually in the Seldon Request Logger context. Each request
    will be attempted for the number of retries, and will return the string
    serialised result.
    Paramters
    ---
    batch_idx
        The enumerated index given to the batch datapoint in order of local dataset
    batch_instance_id
        The unique ID of the batch datapoint created with the python uuid function
    input_raw
        The raw input in string format to be loaded to the respective format
    data_type
        The data type to send which can be str, json and data
    sc
        The instance of SeldonClient to use to send the requests to the seldon model
    retries
        The number of times to retry the request
    batch_id
        The unique identifier for the batch which is passed to all requests
    Returns
    ---
        A string serialised result of the response (or equivallent data with error info)
    """

    predict_kwargs = {}
    meta = {
        "tags": {
            "batch_id": batch_id,
            "batch_instance_id": batch_instance_id,
            "batch_index": batch_idx,
        }
    }
    predict_kwargs["meta"] = meta
    predict_kwargs["headers"] = {"Seldon-Puid": batch_instance_id}
    try:
        data = json.loads(input_raw)
        if data_type == "data":
            data_np = np.array(data)
            predict_kwargs["data"] = data_np
        elif data_type == "str":
            predict_kwargs["str_data"] = data
        elif data_type == "json":
            predict_kwargs["json_data"] = data

        str_output = None
        for i in range(retries):
            try:
                seldon_payload = sc.predict(**predict_kwargs)
                assert seldon_payload.success
                str_output = json.dumps(seldon_payload.response)
                break
            except (requests.exceptions.RequestException, AssertionError) as e:
                logger.error(f"Exception: {e}, retries {retries}")
                if i == (retries - 1):
                    raise

    except Exception as e:
        error_resp = {
            "status": {"info": "FAILURE", "reason": str(e), "status": 1},
            "meta": meta,
        }
        str_output = json.dumps(error_resp)

    return str_output

def _send_batch_feedback(
        input_raw: [],
        sc: SeldonClient,
        retries: int,
        batch_id: str,
) -> str:
    """
    Send an request using the Seldon Client with feedback

    Parameters
    ---
    batch_idx
        The enumerated index given to the batch datapoint in order of local dataset
    batch_instance_id
        The unique ID of the batch datapoint created with the python uuid function
    input_raw
        The raw input in string format to be loaded to the respective format
    data_type
        The data type to send which can be str, json and data
    sc
        The instance of SeldonClient to use to send the requests to the seldon model
    retries
        The number of times to retry the request
    batch_id
        The unique identifier for the batch which is passed to all requests

    Returns
    ---
        A string serialised result of the response (or equivalent data with error info)
    """

    feedback_kwargs = {}
    meta = {
        "tags": {
            "batch_id": batch_id,
            # TODO: tidy these
            "batch_instance_id": input_raw[0][1],
            "batch_index": input_raw[0][0],
        }
    }
    # Feedback protos do not support meta - defined to include in file output only.
    try:
        data = json.loads(input_raw[0][2])
        feedback_kwargs["raw_request"] = data

        str_output = None
        for i in range(retries):
            try:
                seldon_payload = sc.feedback(**feedback_kwargs)
                assert seldon_payload.success

                # Update Tags so we can track feedback instances in output file
                tags = seldon_payload.response.get("meta", {}).get("tags", {})
                tags.update(meta["tags"])
                if "meta" not in seldon_payload.response:
                    seldon_payload.response["meta"] = {}
                seldon_payload.response["meta"]["tags"] = tags
                str_output = json.dumps(seldon_payload.response)
                break
            except (requests.exceptions.RequestException, AssertionError) as e:
                logger.error(f"Exception: {e}, retries {retries}")
                if i == (retries - 1):
                    raise

    except Exception as e:
        error_resp = {
            "status": {"info": "FAILURE", "reason": str(e), "status": 1},
            "meta": meta,
        }
        str_output = json.dumps(error_resp)

    return str_output


@click.command()
@click.option(
    "--deployment-name",
    "-d",
    envvar="SELDON_BATCH_DEPLOYMENT_NAME",
    required=True,
    help="The name of the SeldonDeployment to send the requests to",
)
@click.option(
    "--gateway-type",
    "-g",
    envvar="SELDON_BATCH_GATEWAY_TYPE",
    type=click.Choice(CHOICES_GATEWAY_TYPE),
    default="istio",
    help="The gateway type for the seldon model, which can be through the ingress provider (istio/ambassador) or directly through the service (seldon)",
)
@click.option(
    "--namespace",
    "-n",
    envvar="SELDON_BATCH_NAMESPACE",
    default="default",
    help="The Kubernetes namespace where the SeldonDeployment is deployed in",
)
@click.option(
    "--host",
    "-h",
    envvar="SELDON_BATCH_HOST",
    default="istio-ingressgateway.istio-system.svc.cluster.local:80",
    help="The hostname for the seldon model to send the request to, which can be the ingress of the Seldon model or the service itself",
)
@click.option(
    "--transport",
    "-t",
    envvar="SELDON_BATCH_TRANSPORT",
    type=click.Choice(CHOICES_TRANSPORT),
    default="rest",
    help="The transport type of the SeldonDeployment model which can be REST or GRPC",
)
@click.option(
    "--data-type",
    "-a",
    envvar="SELDON_BATCH_DATA_TYPE",
    type=click.Choice(CHOICES_DATA_TYPE),
    default="data",
    help="Whether to use json, strData or Seldon Data type for the payload to send to the SeldonDeployment which aligns with the SeldonClient format",
)
@click.option(
    "--payload-type",
    "-p",
    envvar="SELDON_BATCH_PAYLOAD_TYPE",
    type=click.Choice(CHOICES_PAYLOAD_TYPE),
    default="ndarray",
    help="The payload type expected by the SeldonDeployment and hence the expected format for the data in the input file which can be an array",
)
@click.option(
    "--workers",
    "-w",
    envvar="SELDON_BATCH_WORKERS",
    type=int,
    default=1,
    help="The number of parallel request processor workers to run for parallel processing",
)
@click.option(
    "--retries",
    "-r",
    envvar="SELDON_BATCH_RETRIES",
    type=int,
    default=3,
    help="The number of retries for each request before marking an error",
)
@click.option(
    "--input-data-path",
    "-i",
    envvar="SELDON_BATCH_INPUT_DATA_PATH",
    type=click.Path(),
    default="/assets/input-data.txt",
    help="The local filestore path where the input file with the data to process is located",
)
@click.option(
    "--output-data-path",
    "-o",
    envvar="SELDON_BATCH_OUTPUT_DATA_PATH",
    type=click.Path(),
    default="/assets/input-data.txt",
    help="The local filestore path where the output file should be written with the outputs of the batch processing",
)
@click.option(
    "--method",
    "-m",
    envvar="SELDON_BATCH_METHOD",
    type=click.Choice(CHOICES_METHOD),
    default="predict",
    help="The method of the SeldonDeployment to send the request to which currently only supports the predict method",
)
@click.option(
    "--log-level",
    "-l",
    envvar="SELDON_BATCH_LOG_LEVEL",
    type=click.Choice(list(CHOICES_LOG_LEVEL)),
    default="info",
    help="The log level for the batch processor",
)
@click.option(
    "--benchmark",
    "-b",
    envvar="SELDON_BATCH_BENCHMARK",
    is_flag=True,
    help="If true the batch processor will print the elapsed time taken to run the process",
)
@click.option(
    "--batch-id",
    "-u",
    envvar="SELDON_BATCH_ID",
    default=str(uuid.uuid1()),
    type=str,
    help="Unique batch ID to identify all datapoints processed in this batch, if not provided is auto generated",
)
@click.option(
    "--batch-size",
    "-u",
    envvar="SELDON_BATCH_SIZE",
    default=int(1),
    type=int,
    help="Batch size greater than 1 can be used to group multiple predictions into a single request.",
)
def run_cli(
        deployment_name: str,
        gateway_type: str,
        namespace: str,
        host: str,
        transport: str,
        data_type: str,
        payload_type: str,
        workers: int,
        retries: int,
        batch_size: int,
        input_data_path: str,
        output_data_path: str,
        method: str,
        log_level: str,
        benchmark: bool,
        batch_id: str,
):
    """
    Command line interface for Seldon Batch Processor, which can be used to send requests
    through configurable parallel workers to Seldon Core models. It is recommended that the
    respective Seldon Core model is also optimized with number of replicas to distribute
    and scale out the batch processing work. The processor is able to process data from local
    filestore input file in various formats supported by the SeldonClient module. It is also
    suggested to use the batch processor component integrated with an ETL Workflow Manager
    such as Kubeflow, Argo Pipelines, Airflow, etc. which would allow for extra setup / teardown
    steps such as downloading the data from object store or starting a seldon core model with replicas.
    See the Seldon Core examples folder for implementations of this batch module with Seldon Core.
    """
    start_multithreaded_batch_worker(
        deployment_name,
        gateway_type,
        namespace,
        host,
        transport,
        data_type,
        payload_type,
        workers,
        retries,
        batch_size,
        input_data_path,
        output_data_path,
        method,
        log_level,
        benchmark,
        batch_id,
    )


if __name__ == '__main__':
    run_cli()
