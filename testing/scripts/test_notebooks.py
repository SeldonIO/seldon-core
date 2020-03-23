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
        raise e


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
    # Misc
    #

    def test_tracing(self):
        create_and_run_script("../../examples/models/tracing", "tracing")

    def test_metrics(self):
        create_and_run_script("../../examples/models/metrics", "metrics")

    def test_payload_logging(self):
        create_and_run_script(
            "../../examples/models/payload_logging", "payload_logging"
        )

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
