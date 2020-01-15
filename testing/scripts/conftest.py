import pytest
from seldon_e2e_utils import clean_string, retry_run, get_s2i_python_version
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


@pytest.fixture
def namespace(request):
    """
    Creates an individual Kubernetes namespace for this particular test and it
    removes it at the end. The value of the injected argument into the test
    function will contain the namespace name.
    """

    test_name = request.node.name
    namespace = clean_string(test_name)

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

    # Install past version cluster-wide
    retry_run("helm delete seldon -n seldon-system")
    retry_run(
        "helm install seldon "
        "seldonio/seldon-core-operator "
        "--namespace seldon-system "
        f"--version {seldon_version} "
        "--wait"
    )

    yield seldon_version

    # Re-install source code version cluster-wide
    retry_run("helm delete seldon -n seldon-system")
    retry_run(
        "helm install seldon "
        "../../helm-charts/seldon-core-operator "
        "--namespace seldon-system "
        "--set istio.enabled=true "
        "--set istio.gateway=seldon-gateway "
        "--set certManager.enabled=false "
        "--wait",
        attempts=2,
    )


def do_s2i_python_version():
    return get_s2i_python_version()
