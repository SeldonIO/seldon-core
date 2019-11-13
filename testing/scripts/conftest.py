import pytest
from seldon_e2e_utils import get_s2i_python_version
from subprocess import run


@pytest.fixture(scope="session", autouse=True)
def run_pod_information_in_background(request):
    # This command runs the pod information and prints it
    #   in the background every time there's a new update
    run(
        (
            "kubectl get pods --all-namespaces -w | "
            + 'awk \'{print "\\nPODS UPDATE: "$0"\\n"}\' & '
        ),
        shell=True,
    )


@pytest.fixture(scope="module")
def s2i_python_version():
    return do_s2i_python_version()


def do_s2i_python_version():
    return get_s2i_python_version()
