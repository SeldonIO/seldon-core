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
    result_body += f"* All mean performance latency under 5ms: {latency_mean}"
    # Ensure 99th percentiles are not spiking above 15ms
    latency_nth = all(df["99th"] < 10)
    result_body += f"* All 99th performance latenc under 10ms: {latency_nth}"
    # Ensure throughput is above 200 rps for REST
    rps_rest = all(df[df["apiType"] == "rest"]["throughputAchieved"] > 200)
    result_body += f"* REST throughput above 200rps: {rps_rest}"
    # Ensure throughput is above 250 rps for GRPC
    rps_grpc = all(df[df["apiType"] == "grpc"]["throughputAchieved"] > 250)
    result_body += f"* GRPC throughput above 250rps: {rps_grpc}"
    # Validate latenc added by adding service orchestrator is lower than 4ms
    orch_mean = all(
        (
            df[df["disableOrchestrator"] == "true"]["mean"].values
            - df[df["disableOrchestrator"] == "false"]["mean"].values
        )
        < 2
    )
    result_body += f"* Orch added mean latency under 2ms: {orch_mean}"
    orch_nth = all(
        (
            df[df["disableOrchestrator"] == "true"]["99th"].values
            - df[df["disableOrchestrator"] == "false"]["99th"].values
        )
        < 2
    )
    result_body += f"* Orch added 99th latency under 2ms: {orch_nth}"

    result_body += "\n### Results table\n\n"
    result_body += str(df.to_markdown())
    post_comment_in_pr(result_body)

    assert latency_mean
    assert latency_nth
    assert rps_rest
    assert rps_grpc
    assert orch_mean
    assert orch_nth


