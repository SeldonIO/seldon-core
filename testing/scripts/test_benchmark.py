import json
import logging

import pandas as pd
import pytest

from seldon_e2e_utils import post_comment_in_pr, run_benchmark_and_capture_results


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_service_orchestrator():

    df = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        disable_orchestrator_list=["false", "true"],
    )

    result_body = "# Benchmark results\n\n"

    # Ensure all mean performance latency below 5 ms
    latency_mean = all(df["mean"] < 5)
    result_body += f"* All mean performance latency under 5ms: {latency_mean}\n"
    # Ensure 99th percentiles are not spiking above 15ms
    latency_nth = all(df["99th"] < 10)
    result_body += f"* All 99th performance latenc under 10ms: {latency_nth}\n"
    # Ensure throughput is above 200 rps for REST
    rps_rest = all(df[df["apiType"] == "rest"]["throughputAchieved"] > 200)
    result_body += f"* REST throughput above 200rps: {rps_rest}\n"
    # Ensure throughput is above 250 rps for GRPC
    rps_grpc = all(df[df["apiType"] == "grpc"]["throughputAchieved"] > 250)
    result_body += f"* GRPC throughput above 250rps: {rps_grpc}\n"
    # Validate latenc added by adding service orchestrator is lower than 4ms
    orch_mean = all(
        (
            df[df["disableOrchestrator"] == "true"]["mean"].values
            - df[df["disableOrchestrator"] == "false"]["mean"].values
        )
        < 2
    )
    result_body += f"* Orch added mean latency under 2ms: {orch_mean}\n"
    orch_nth = all(
        (
            df[df["disableOrchestrator"] == "true"]["99th"].values
            - df[df["disableOrchestrator"] == "false"]["99th"].values
        )
        < 2
    )
    result_body += f"* Orch added 99th latency under 2ms: {orch_nth}\n"

    result_body += "\n### Results table\n\n"
    result_body += str(df.to_markdown())
    post_comment_in_pr(result_body)

    assert latency_mean
    assert latency_nth
    assert rps_rest
    assert rps_grpc
    assert orch_mean
    assert orch_nth


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_python_wrapper_v1_vs_v2():

    result_body = ""
    result_body += "\n# Benchmark Python Wrapper V1 vs V2\n\n"

    df_pywrapper = run_benchmark_and_capture_results(
       api_type_list=["rest", "grpc"],
       protocol="seldon",
       server_list=["SKLEARN_SERVER"],
       benchmark_data={"data": {"ndarray": [[1, 2, 3, 4]]}},
    )

    result_body += "\n### Python V1 Wrapper Results table\n\n"
    result_body += str(df_pywrapper.to_markdown())

    # TODO: Validate equivallent of parallel workers in MLServer
    df_mlserver = run_benchmark_and_capture_results(
        api_type_list=["rest", "grpc"],
        model_name="classifier",
        protocol="kfserving",
        server_list=["SKLEARN_SERVER"],
        model_uri_list=["gs://seldon-models/sklearn/iris-0.23.2/lr_model"],
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

    result_body += "\n\n\n### Python V2 MLServer Results table\n\n"
    result_body += str(df_mlserver.to_markdown())

    post_comment_in_pr(result_body)
