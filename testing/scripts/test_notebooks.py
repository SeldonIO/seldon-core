from subprocess import run, PIPE, CalledProcessError
import logging
import pytest


def create_and_run_script(folder, notebook):
    run(
        f"jupyter nbconvert --template ../../notebooks/convert.tpl --to script {folder}/{notebook}.ipynb",
        shell=True,
        check=True,
    )
    run(f"chmod u+x {folder}/{notebook}.py", shell=True, check=True)
    try:
        run(
            f"cd {folder} && ./{notebook}.py",
            shell=True,
            check=True,
            stdout=PIPE,
            stderr=PIPE,
            encoding="utf-8",
        )
    except CalledProcessError as e:
        logging.error(
            f"failed notebook test {notebook} stdout:{e.stdout}, stderr:{e.stderr}"
        )
        run("kubectl delete sdep --all", shell=True, check=False)
        raise e


@pytest.mark.flaky(max_runs=2)
@pytest.mark.notebooks
class TestNotebooks(object):

    #
    # Core notebooks
    #

    def test_helm_examples(self):
        create_and_run_script("../../notebooks", "helm_examples")

    def test_explainer_examples(self):
        create_and_run_script("../../notebooks", "explainer_examples")

    def test_istio_examples(self):
        create_and_run_script("../../notebooks", "istio_example")

    def test_max_grpc_msg_size(self):
        create_and_run_script("../../notebooks", "max_grpc_msg_size")

    def test_multiple_operators(self):
        create_and_run_script("../../notebooks", "multiple_operators")

    def test_protocol_examples(self):
        create_and_run_script("../../notebooks", "protocol_examples")

    def test_server_examples(self):
        create_and_run_script("../../notebooks", "server_examples")

    def test_rolling_updates(self):
        create_and_run_script("../../notebooks", "rolling_updates")

    #
    # Ambassador
    #

    def test_ambassador_canary(self):
        create_and_run_script("../../examples/ambassador/canary", "ambassador_canary")

    def test_ambassador_headers(self):
        create_and_run_script("../../examples/ambassador/headers", "ambassador_headers")

    def test_ambassador_shadow(self):
        create_and_run_script("../../examples/ambassador/shadow", "ambassador_shadow")

    def test_ambassador_custom(self):
        create_and_run_script("../../examples/ambassador/custom", "ambassador_custom")

    #
    # Istio Examples
    #

    def test_istio_canary(self):
        create_and_run_script("../../examples/istio/canary_update", "canary")

    #
    # KEDA Examples
    #

    def test_keda_prom_auto_scale(self):
        try:
            create_and_run_script("../../examples/keda", "keda_prom_auto_scale")
        except CalledProcessError as e:
            run(
                f"helm delete seldon-core-analytics --namespace seldon-system",
                shell=True,
                check=False,
            )
            raise e

    #
    # Misc
    #

    def test_tracing(self):
        create_and_run_script("../../examples/models/tracing", "tracing")

    def test_metrics(self):
        try:
            create_and_run_script("../../examples/models/metrics", "metrics")
        except CalledProcessError as e:
            run(
                f"helm delete seldon-core-analytics --namespace seldon-system",
                shell=True,
                check=False,
            )
            raise e

    def test_metadata(self):
        create_and_run_script("../../examples/models/metadata", "metadata")

    def test_graph_metadata(self):
        create_and_run_script("../../examples/models/metadata", "graph_metadata")

    def test_grpc_metadata(self):
        create_and_run_script("../../examples/models/metadata", "metadata_grpc")

    def test_payload_logging(self):
        create_and_run_script(
            "../../examples/models/payload_logging", "payload_logging"
        )

    def test_custom_metrics(self):
        try:
            create_and_run_script(
                "../../examples/models/custom_metrics", "customMetrics"
            )
        except CalledProcessError as e:
            run(
                f"helm delete seldon-core-analytics --namespace seldon-system",
                shell=True,
                check=False,
            )
            raise e

    def test_autoscaling(self):
        try:
            create_and_run_script(
                "../../examples/models/autoscaling", "autoscaling_example"
            )
        except CalledProcessError as e:
            run(f"helm delete loadtester --namespace seldon", shell=True, check=False)
            raise e

    def test_scaling(self):
        create_and_run_script("../../notebooks", "scale")

    #
    # Payloads
    #

    def test_jsondata(self):
        create_and_run_script(
            "../../examples/models/sklearn_iris_jsondata", "sklearn_iris_jsondata"
        )

    #
    # SKLearn
    #

    def test_sklearn_iris(self):
        create_and_run_script("../../examples/models/sklearn_iris", "sklearn_iris")

    #
    # OpenVino
    #

    # def test_openvino_squeezenet(self):
    #    create_and_run_script("../../examples/models/openvino", "openvino-squeezenet")

    # def test_openvino_imagenet_ensemble(self):
    #    create_and_run_script(
    #        "../../examples/models/openvino_imagenet_ensemble",
    #        "openvino_imagenet_ensemble",
    #    )

    def test_custom_metrics_server(self):
        create_and_run_script("../../examples/feedback/metrics-server", "README")

    #
    # Upgrade
    #

    def test_upgrade(self):
        try:
            create_and_run_script("../../notebooks", "operator_upgrade")
        except:
            run("make install_seldon", shell=True, check=False)
            raise

    def test_disruption_budgets(self):
        create_and_run_script(
            "../../examples/models/disruption_budgets", "pdbs_example"
        )
