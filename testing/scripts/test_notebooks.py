from subprocess import run

def create_and_run_script(notebook):
    run(
        f"jupyter nbconvert --template ../../notebooks/convert.tpl --to script ../../notebooks/{notebook}.ipynb",
        shell=True,
        check=True,
    )
    run(f"chmod u+x ../../notebooks/{notebook}.py", shell=True, check=True)
    run(f"cd ../../notebooks && ./{notebook}.py", shell=True, check=True)

class TestNotebooks(object):

    def test_helm_examples(self):
        create_and_run_script("helm_examples")

    def test_explainer_examples(self):
        create_and_run_script("explainer_examples")

    def test_istio_examples(self):
        create_and_run_script("istio_example")

    def test_max_grpc_msg_size(self):
        create_and_run_script("max_grpc_msg_size")

    def test_multiple_operators(self):
        create_and_run_script("multiple_operators")

    def test_protocol_examples(self):
        create_and_run_script("protocol_examples")

    def test_server_examples(self):
        create_and_run_script("server_examples")