import json

import numpy as np
import pytest
import tensorflow as tf
from google.protobuf import json_format

from seldon_e2e_utils import post_comment_in_pr, run_benchmark_and_capture_results


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_service_orchestrator():

    sort_by = ["apiType", "disableOrchestrator"]

    data_size = 1_000
    data = [100.0] * data_size

    data_tensor = {"data": {"tensor": {"values": data, "shape": [1, data_size]}}}

    df = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        disable_orchestrator_list=["false", "true"],
        image_list=["seldonio/seldontest_predict:1.10.0-dev"],
        benchmark_data=data_tensor,
    )
    df = df.sort_values(sort_by)

    result_body = "# Benchmark results - Testing Service Orchestrator\n\n"

    orch_mean = all(
        (
            df[df["disableOrchestrator"] == "false"]["mean"].values
            - df[df["disableOrchestrator"] == "true"]["mean"].values
        )
        < 3
    )
    result_body += f"* Orch added mean latency under 4ms: {orch_mean}\n"
    orch_nth = all(
        (
            df[df["disableOrchestrator"] == "false"]["95th"].values
            - df[df["disableOrchestrator"] == "true"]["95th"].values
        )
        < 5
    )
    result_body += f"* Orch added 95th latency under 5ms: {orch_nth}\n"
    orch_nth = all(
        (
            df[df["disableOrchestrator"] == "false"]["99th"].values
            - df[df["disableOrchestrator"] == "true"]["99th"].values
        )
        < 10
    )
    result_body += f"* Orch added 99th latency under 10ms: {orch_nth}\n"

    # We have to set no errors to 1 as the tools for some reason have 1 as base
    no_err = all(df["errors"] <= 1)
    result_body += f"* No errors: {no_err}\n"

    result_body += "\n### Results table\n\n"
    result_body += str(df.to_markdown())
    post_comment_in_pr(result_body)

    assert orch_mean
    assert orch_nth


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_workers_performance():

    sort_by = ["apiType", "serverWorkers"]

    data_size = 10
    data = [100.0] * data_size

    data_tensor = {"data": {"tensor": {"values": data, "shape": [1, data_size]}}}

    df = run_benchmark_and_capture_results(
        api_type_list=["grpc", "rest"],
        server_workers_list=["1", "5", "10"],
        benchmark_concurrency_list=["10", "100", "1000"],
        parallelism="1",
        requests_cpu_list=["4000Mi"],
        limits_cpu_list=["4000Mi"],
        image_list=["seldonio/seldontest_predict:1.10.0-dev"],
        benchmark_data=data_tensor,
    )
    df = df.sort_values(sort_by)

    result_body = "# Benchmark results - Testing Workers Performance\n\n"

    result_body += "\n### Results table\n\n"
    result_body += str(df.to_markdown())
    post_comment_in_pr(result_body)


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_python_wrapper_v1_vs_v2_iris():

    sort_by = ["concurrency", "apiType"]
    benchmark_concurrency_list = ["1", "50", "150"]

    result_body = ""
    result_body += "\n# Benchmark Results - Python Wrapper V1 vs V2\n\n"

    # Using single worker as fastapi also uses single worker
    df_pywrapper = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        protocol="seldon",
        server_list=["SKLEARN_SERVER"],
        benchmark_concurrency_list=benchmark_concurrency_list,
        model_uri_list=["gs://seldon-models/v1.11.0-dev/sklearn/iris"],
        benchmark_data={"data": {"ndarray": [[1, 2, 3, 4]]}},
    )
    df_pywrapper = df_pywrapper.sort_values(sort_by)

    conc_idx = df_pywrapper["concurrency"] == 1
    # Python V1 Wrapper Validations
    # Ensure base mean performance latency below 10 ms
    v1_latency_mean = all((df_pywrapper[conc_idx]["mean"] < 10))
    result_body += f"* V1 base mean performance latency under 10ms: {v1_latency_mean}\n"
    # Ensure 99th percentiles are not spiking above 15ms
    v1_latency_nth = all(df_pywrapper[conc_idx]["99th"] < 10)
    result_body += f"* V1 base 99th performance latenc under 10ms: {v1_latency_nth}\n"
    # Ensure throughput is above 180 rps for REST
    v1_rps_rest = all(
        df_pywrapper[(df_pywrapper["apiType"] == "rest") & conc_idx][
            "throughputAchieved"
        ]
        > 180
    )
    result_body += f"* V1 base throughput above 180rps: {v1_rps_rest}\n"
    # Ensure throughput is above 250 rps for GRPC
    v1_rps_grpc = all(
        df_pywrapper[(df_pywrapper["apiType"] == "grpc") & conc_idx][
            "throughputAchieved"
        ]
        > 250
    )
    result_body += f"* V1 base throughput above 250rps: {v1_rps_grpc}\n"
    # Validate latenc added by adding service orchestrator is lower than 4ms

    # TODO: Validate equivallent of parallel workers in MLServer
    df_mlserver = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        model_name="classifier",
        protocol="kfserving",
        server_list=["SKLEARN_SERVER"],
        model_uri_list=["gs://seldon-models/sklearn/iris-0.23.2/lr_model"],
        benchmark_concurrency_list=benchmark_concurrency_list,
        benchmark_data={
            "inputs": [
                {
                    "name": "predict",
                    "datatype": "FP32",
                    "shape": [1, 4],
                    "data": [[1, 2, 3, 4]],
                }
            ]
        },
        benchmark_grpc_data_override={
            "model_name": "classifier",
            "inputs": [
                {
                    "name": "predict",
                    "datatype": "FP32",
                    "shape": [1, 4],
                    "contents": {"fp32_contents": [1, 2, 3, 4]},
                }
            ],
        },
    )
    # First we sort the dataframes to ensure they are compared correctly
    df_mlserver = df_mlserver.sort_values(sort_by)

    # Python V1 Wrapper Validations

    conc_idx = df_mlserver["concurrency"] == 1
    # Ensure all mean performance latency below 5 ms
    v2_latency_mean = all(df_mlserver[conc_idx]["mean"] < 5)
    result_body += f"* V2 mean performance latency under 5ms: {v2_latency_mean}\n"
    # Ensure 99th percentiles are not spiking above 15ms
    v2_latency_nth = all(df_mlserver[conc_idx]["99th"] < 10)
    result_body += f"* V2 99th performance latenc under 10ms: {v2_latency_nth}\n"
    # Ensure throughput is above 180 rps for REST
    v2_rps_rest = all(
        df_mlserver[(df_mlserver["apiType"] == "rest") & conc_idx]["throughputAchieved"]
        > 250
    )
    result_body += f"* V2 REST throughput above 250rps: {v2_rps_rest}\n"
    # Ensure throughput is above 250 rps for GRPC
    v2_rps_grpc = all(
        df_mlserver[(df_mlserver["apiType"] == "grpc") & conc_idx]["throughputAchieved"]
        > 250
    )
    result_body += f"* V2 throughput above 300rps: {v2_rps_grpc}\n"

    result_body += "\n### Python V1 Wrapper Results table\n\n"
    result_body += str(df_pywrapper.to_markdown())
    result_body += "\n\n\n### Python V2 MLServer Results table\n\n"
    result_body += str(df_mlserver.to_markdown())

    post_comment_in_pr(result_body)

    assert v1_latency_mean
    assert v1_latency_nth
    assert v1_rps_rest
    assert v1_rps_grpc
    assert v2_latency_mean
    assert v2_latency_nth
    assert v2_rps_rest
    assert v2_rps_grpc


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_v1_seldon_data_types():

    sort_by = ["concurrency", "apiType"]

    # 10000 element array
    data_size = 10_000
    data = [100.0] * data_size

    benchmark_concurrency_list = ["1", "50", "150"]

    image_list = ["seldonio/seldontest_predict:1.10.0-dev"]

    data_ndarray = {"data": {"ndarray": data}}
    data_tensor = {"data": {"tensor": {"values": data, "shape": [1, data_size]}}}

    array = np.array(data)
    tftensor_proto = tf.make_tensor_proto(array)
    tftensor_json_str = json_format.MessageToJson(tftensor_proto)
    tftensor_dict = json.loads(tftensor_json_str)
    data_tftensor = {"data": {"tftensor": tftensor_dict}}

    df_ndarray = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        image_list=image_list,
        benchmark_concurrency_list=benchmark_concurrency_list,
        benchmark_data=data_ndarray,
    )
    df_ndarray = df_ndarray.sort_values(sort_by)

    df_tensor = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        image_list=image_list,
        benchmark_concurrency_list=benchmark_concurrency_list,
        benchmark_data=data_tensor,
    )
    df_tensor = df_tensor.sort_values(sort_by)

    df_tftensor = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        image_list=image_list,
        benchmark_concurrency_list=benchmark_concurrency_list,
        benchmark_data=data_tftensor,
    )
    df_tftensor = df_tftensor.sort_values(sort_by)

    result_body = "# Benchmark results - Testing Seldon V1 Data Types\n\n"

    result_body += "\n### Results for NDArray\n\n"
    result_body += str(df_ndarray.to_markdown())
    result_body += "\n### Results for Tensor\n\n"
    result_body += str(df_tensor.to_markdown())
    result_body += "\n### Results for TFTensor\n\n"
    result_body += str(df_tftensor.to_markdown())
    post_comment_in_pr(result_body)
