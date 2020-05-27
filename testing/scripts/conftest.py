import os
import pytest
from seldon_e2e_utils import clean_string, retry_run, get_seldon_version

from e2e_utils.install import install_seldon, delete_seldon
from subprocess import run


def _to_python_bool(val):
    # From Flask's docs:
    # https://flask.palletsprojects.com/en/1.1.x/config/#configuring-from-environment-variables
    return val.lower() in {"1", "t", "true"}


SELDON_E2E_TESTS_USE_EXECUTOR = _to_python_bool(
    os.getenv("SELDON_E2E_TESTS_USE_EXECUTOR", default="true")
)
SELDON_E2E_TESTS_POD_INFORMATION = _to_python_bool(
    os.getenv("SELDON_E2E_TESTS_POD_INFORMATION", default="false")
)


@pytest.fixture(scope="session", autouse=SELDON_E2E_TESTS_POD_INFORMATION)
def run_pod_information_in_background(request):
    # This command runs the pod information and prints it in the background
    # every time there's a new update
    run(
        (
            "kubectl get pods --all-namespaces -w | "
            + 'awk \'{print "\\nPODS UPDATE: "$0"\\n"}\' & '
        ),
        shell=True,
    )


@pytest.fixture(scope="module")
def s2i_python_version():
    """Return version of s2i images, the IMAGE_VERSION, e.g. 0.17-SNAPSHOT."""
    return do_s2i_python_version()


@pytest.fixture
def namespace(request):
    """
    Creates an individual Kubernetes namespace for this particular test and it
    removes it at the end. The value of the injected argument into the test
    function will contain the namespace name.
    """

    test_name = request.node.name
    namespace = clean_string(test_name)
    if len(namespace) > 63:
        namespace = namespace[0:63]

    # Create namespace
    retry_run(f"kubectl create namespace {namespace}")
    yield namespace

    # Tear down namespace
    run(f"kubectl delete namespace {namespace}", shell=True)


@pytest.fixture
def seldon_version(request):
    """
    Ensures that the cluster-wide version of Seldon Core is set to a particular
    version. After the test finishes it restores the code base version
    (SNAPSHOT).

    Since this fixture needs a parameter (the version we want on the
    cluster), it needs to be used as an indirect parametrization.

    ```python
        @pytest.mark.parametrize("seldon_version", ["1.0.0", "1.0.1"], indirect=True)
        def test_old_versions(seldon_version):
            ...
    ```
    """
    seldon_version = request.param

    # Delete source version cluster-wide and install new one
    delete_seldon()
    install_seldon(version=seldon_version, executor=SELDON_E2E_TESTS_USE_EXECUTOR)

    yield seldon_version

    # Re-install source code version cluster-wide
    delete_seldon()
    install_seldon(executor=SELDON_E2E_TESTS_USE_EXECUTOR)


def do_s2i_python_version():
    return get_seldon_version()
