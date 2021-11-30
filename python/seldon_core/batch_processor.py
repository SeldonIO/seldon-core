import copy
import json
import logging
import time
import uuid
from itertools import groupby
from queue import Empty, Queue
from threading import Event, Thread
from typing import Dict, List, Tuple

import click
import numpy as np
import requests

from seldon_core.seldon_client import SeldonClient, SeldonCallCredentials

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
    batch_interval: float,
    call_credentials_token: str,
    use_ssl: bool,
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
    elif data_type not in ["data", "raw"] and batch_size > 1:
        raise RuntimeError(
            "Batch size greater than 1 is only supported for `data` data type."
        )

    # Providing call credentials sets the REST transport protocol to https,
    # so we configure credentials even without a supplied token is use_ssl is set.
    credentials = None
    if use_ssl or len(call_credentials_token) > 0:
        token = None
        if len(call_credentials_token) > 0:
            token = call_credentials_token
        credentials = SeldonCallCredentials(token=token)

    sc = SeldonClient(
        gateway=gateway_type,
        transport=transport,
        deployment_name=deployment_name,
        payload_type=payload_type,
        gateway_endpoint=host,
        namespace=namespace,
        client_return_type="dict",
        call_credentials=credentials,
    )

    t_in = Thread(
        target=_start_input_file_worker,
        args=(q_in, input_data_path, batch_size),
        daemon=True,
    )
    t_in.start()

    for _ in range(workers):
        Thread(
            target=_start_request_worker,
            args=(
                q_in,
                q_out,
                data_type,
                sc,
                method,
                retries,
                batch_id,
                payload_type,
                batch_interval,
            ),
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
        logger.debug(f"Elapsed time: {time.time() - start_time}")


def _start_input_file_worker(
    q_in: Queue, input_data_path: str, batch_size: int
) -> None:
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
    if batch:
        q_in.put(batch)


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
    batch_interval: float,
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
        The json/str/data/raw type to send the requests as
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
        start_time = time.time()
        input_data = q_in.get()
        if method == "predict":
            # If we have a batch size > 1 then we wish to use the method for sending multiple predictions
            # as a single request and split the response into multiple responses.
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
            else:
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
                q_out.put(str_output)
        elif method == "feedback":
            batch_idx, batch_instance_id, input_raw = input_data[0]
            str_output = _send_batch_feedback(
                batch_idx,
                batch_instance_id,
                input_raw,
                data_type,
                sc,
                retries,
                batch_id,
            )
            q_out.put(str_output)

        # Setting time interval before the task is marked as done
        if batch_interval > 0:
            remaining_interval = batch_interval - (time.time() - start_time)
            if remaining_interval > 0:
                time.sleep(remaining_interval)

        # Mark task as done in the queue to add space for new tasks
        q_in.task_done()


def _extract_raw_data_multi_request(
    loaded_data: List[Dict], tags: Dict
) -> Tuple[Dict, str, Dict]:
    raw_input_tags = [d.get("meta", {}).get("tags", {}) for d in loaded_data]
    first_input = loaded_data[0]

    # Raw input format in mini-batch mode only work for "data" format
    if "data" not in first_input:
        raise ValueError(
            "raw input with predict in mini-batch mode requires data payload"
        )
    # If-block for ndarray case
    elif "ndarray" in first_input["data"]:
        payload_type = "ndarray"
        names_list = [d["data"]["names"] for d in loaded_data]
        arrays = [np.array(d["data"]["ndarray"]) for d in loaded_data]
        if not all(names_list[0] == name for name in names_list):
            raise ValueError("All names in mini-batch must be the same.")
        for arr in arrays:
            if arr.shape[0] != 1:
                raise ValueError(
                    "When using mini-batching each row should contain single instance."
                )
        ndarray = np.concatenate(arrays)
        raw_data = {
            "data": {"names": names_list[0], "ndarray": ndarray.tolist()},
            "meta": {"tags": tags},
        }
        return raw_data, payload_type, raw_input_tags

    # If-block for tensor case
    elif "tensor" in first_input["data"]:
        payload_type = "tensor"
        names_list = [d["data"]["names"] for d in loaded_data]
        tensor_shapes = [d["data"]["tensor"]["shape"] for d in loaded_data]
        tensor_values = [d["data"]["tensor"]["values"] for d in loaded_data]

        if not all(names_list[0] == name for name in names_list):
            raise ValueError("All names in mini-batch must be the same.")

        dim_0 = 0
        dim_1 = tensor_shapes[0][1]
        for shape in tensor_shapes:
            if shape[0] != 1:
                raise ValueError(
                    "When using mini-batching each row should contain single instance."
                )
            dim_0 += shape[0]
            if dim_1 != shape[1]:
                raise ValueError(
                    "All instances in mini-batch must have same number of features."
                )
        values = sum(tensor_values, [])
        shape = [dim_0, dim_1]
        raw_data = {
            "data": {
                "names": names_list[0],
                "tensor": {"shape": shape, "values": values},
            },
            "meta": {"tags": tags},
        }
        return raw_data, payload_type, raw_input_tags


def _send_batch_predict_multi_request(
    input_data: [],
    data_type: str,
    sc: SeldonClient,
    retries: int,
    batch_id: str,
    payload_type: str,
) -> [str]:
    """
    Send an request using the Seldon Client with batch context including the
    unique ID of the batch and the Batch enumerated index as metadata. This
    function also uses the unique batch ID as request ID so the request can be
    traced back individually in the Seldon Request Logger context. Each request
    will be attempted for the number of retries, and will return the string
    serialised result. This method is similar to _send_batch_predict, but allows multiple
    requests to be combined into a single prediction.
    Parameters
    ---
    input_data
        The input data containing the indexes, instance_ids and predictions
    data_type
        The data type to send which can be `data`
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

    indexes = [x[0] for x in input_data]

    seldon_puid = input_data[0][1]
    instance_ids = [f"{seldon_puid}-item-{n}" for n, _ in enumerate(input_data)]
    loaded_data = [json.loads(data[2]) for data in input_data]

    predict_kwargs = {}
    tags = {
        "batch_id": batch_id,
    }
    predict_kwargs["meta"] = tags
    predict_kwargs["headers"] = {"Seldon-Puid": seldon_puid}

    try:
        # Process raw input format
        if data_type == "raw":
            raw_data, payload_type, raw_input_tags = _extract_raw_data_multi_request(
                loaded_data, predict_kwargs["meta"]
            )
            predict_kwargs["raw_data"] = raw_data
        else:
            # Initialise concatenated array for data
            arrays = [np.array(arr) for arr in loaded_data]
            for arr in arrays:
                if arr.shape[0] != 1:
                    raise ValueError(
                        "When using mini-batching each row should contain single instance."
                    )
            concat = np.concatenate(arrays)
            predict_kwargs["data"] = concat
        logger.debug(f"calling sc.predict with {predict_kwargs}")
    except Exception as e:
        error_resp = {
            "status": {"info": "FAILURE", "reason": str(e), "status": 1},
            "meta": tags,
        }
        logger.error(f"Exception: {e}")
        str_output = json.dumps(error_resp)
        return [str_output]

    try:
        for i in range(retries):
            try:
                seldon_payload = sc.predict(**predict_kwargs)
                assert seldon_payload.success
                response = seldon_payload.response
                break
            except (requests.exceptions.RequestException, AssertionError) as e:
                logger.error(
                    f"Exception: {e}, retries {i+1} / {retries} for batch_id(s)={indexes}"
                )
                if i == (retries - 1):
                    raise

    except Exception as e:
        output = []
        for batch_index, batch_instance_id in zip(indexes, instance_ids):
            error_resp = {
                "status": {"info": "FAILURE", "reason": str(e), "status": 1},
                "meta": dict(
                    batch_index=batch_index, batch_instance_id=batch_instance_id, **tags
                ),
            }
            logger.error(f"Exception: {e}")
            output.append(json.dumps(error_resp))
        return output

    # Take the response create new responses for each request
    responses = []

    # If tensor then prepare the ndarray
    if payload_type == "tensor":
        tensor = np.array(response["data"]["tensor"]["values"])
        shape = response["data"]["tensor"]["shape"]
        tensor_ndarray = tensor.reshape(shape)

    for i in range(len(input_data)):
        try:
            new_response = copy.deepcopy(response)
            if data_type == "raw":
                new_response["meta"]["tags"].update(raw_input_tags[i])
            if payload_type == "ndarray":
                # Format new responses for each original prediction request
                new_response["data"]["ndarray"] = [response["data"]["ndarray"][i]]
                new_response["meta"]["tags"]["batch_index"] = indexes[i]
                new_response["meta"]["tags"]["batch_instance_id"] = instance_ids[i]

                responses.append(json.dumps(new_response))
            elif payload_type == "tensor":
                # Format new responses for each original prediction request
                new_response["data"]["tensor"]["shape"][0] = 1
                new_response["data"]["tensor"]["values"] = np.ndarray.tolist(
                    tensor_ndarray[i]
                )
                new_response["meta"]["tags"]["batch_index"] = indexes[i]
                new_response["meta"]["tags"]["batch_instance_id"] = instance_ids[i]
                responses.append(json.dumps(new_response))
            else:
                raise RuntimeError(
                    "Only `ndarray` and `tensor` input are currently supported for batch size greater than 1."
                )
        except Exception as e:
            error_resp = {
                "status": {"info": "FAILURE", "reason": str(e), "status": 1},
                "meta": tags,
            }
            logger.error("Exception: %s" % e)
            responses.append(json.dumps(error_resp))

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
    Parameters
    ---
    batch_idx
        The enumerated index given to the batch datapoint in order of local dataset
    batch_instance_id
        The unique ID of the batch datapoint created with the python uuid function
    input_raw
        The raw input in string format to be loaded to the respective format
    data_type
        The data type to send which can be `str`, `json` and `data`
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

    predict_kwargs = {}
    tags = {
        "batch_id": batch_id,
        "batch_instance_id": batch_instance_id,
        "batch_index": batch_idx,
    }
    predict_kwargs["meta"] = tags
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
        elif data_type == "raw":
            # Make sure data contains meta.tags keys.
            data["meta"] = data.get("meta", {})
            data["meta"]["tags"] = data["meta"].get("tags", {})

            # Update them with our
            data["meta"]["tags"].update(tags)
            predict_kwargs["raw_data"] = data

        logger.debug(f"calling sc.predict with {predict_kwargs}")

        str_output = None
        for i in range(retries):
            try:
                seldon_payload = sc.predict(**predict_kwargs)
                assert seldon_payload.success
                str_output = json.dumps(seldon_payload.response)
                break
            except (requests.exceptions.RequestException, AssertionError) as e:
                logger.error(
                    f"Exception: {e}, retries {i+1} / {retries} for batch_index={batch_idx}"
                )
                if i == (retries - 1):
                    raise

    except Exception as e:
        error_resp = {
            "status": {"info": "FAILURE", "reason": str(e), "status": 1},
            "meta": tags,
        }
        logger.error("Exception: %s" % e)
        str_output = json.dumps(error_resp)

    return str_output


def _send_batch_feedback(
    batch_idx: int,
    batch_instance_id: int,
    input_raw: str,
    data_type: str,
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
            "batch_instance_id": batch_instance_id,
            "batch_index": batch_idx,
        }
    }
    # Feedback protos do not support meta - defined to include in file output only.
    try:
        data = json.loads(input_raw)
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
        logger.error("Exception: %s" % e)
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
    help="Unique batch ID to identify all data points processed in this batch, if not provided is auto generated",
)
@click.option(
    "--batch-size",
    "-s",
    envvar="SELDON_BATCH_SIZE",
    default=1,
    type=int,
    help="Batch size greater than 1 can be used to group multiple predictions into a single request.",
)
@click.option(
    "--batch-interval",
    "-t",
    envvar="SELDON_BATCH_MIN_INTERVAL",
    default=0,
    type=float,
    help="Minimum Time interval (in seconds) between batch predictions made by every worker.",
)
@click.option(
    "--use-ssl",
    envvar="SELDON_BATCH_USE_SSL",
    default=False,
    type=bool,
    help="Whether to use https rather than http as the REST transport protocol.",
)
@click.option(
    "--call-credentials-token",
    envvar="SELDON_BATCH_CALL_CREDENTIALS_TOKEN",
    default="",
    type=str,
    help="Auth token used by Seldon Client, if supplied and using REST the transport protocol will be https.",
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
    batch_interval: float,
    call_credentials_token: str,
    use_ssl: bool,
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
        batch_interval,
        call_credentials_token,
        use_ssl,
    )


if __name__ == "__main__":
    run_cli()
