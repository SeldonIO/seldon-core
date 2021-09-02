import os
from subprocess import run

import pytest

from e2e_utils.install import delete_seldon, install_seldon
from e2e_utils.s2i import create_s2i_image, kind_load_image
from seldon_e2e_utils import clean_string, get_seldon_version, retry_run


def _to_python_bool(val):
    # From Flask's docs:
    # https://flask.palletsprojects.com/en/1.1.x/config/#configuring-from-environment-variables
    return val.lower() in {"1", "t", "true"}


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


def install_argo():
    kwargs = {
        "check": True,
        "shell": True,
    }

    run("kubectl create namespace argo", **kwargs)
    run(
        "kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo-workflows/stable/manifests/install.yaml",
        **kwargs,
    )
    run("kubectl rollout status -n argo deployment/argo-server", **kwargs)
    run("kubectl rollout status -n argo deployment/workflow-controller", **kwargs)
    run(
        "kubectl create rolebinding argo-default-admin --clusterrole=admin --serviceaccount=argo:default -n argo",
        **kwargs,
    )
    run(
        "kubectl create rolebinding argo-seldon-workflow --clusterrole=seldon-manager-role-seldon-system --serviceaccount=argo:default -n argo",
        **kwargs,
    )
    run("kubectl apply -n argo -f ../resources/argo-configmap.yaml", **kwargs)


def delete_argo():
    run("kubectl delete namespace argo", check=True, shell=True)


@pytest.fixture()
def argo_worfklows(scope="module"):
    install_argo()
    yield
    delete_argo()


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
    install_seldon(version=seldon_version)

    yield seldon_version

    # Re-install source code version cluster-wide
    delete_seldon()
    install_seldon()


@pytest.fixture(scope="module")
def s2i_image(request):
    """
    Creates an S2I image.
    Note that this is an indirect fixture, therefore it will read some
    parameters.

    Parameters
    ---
    s2i_folder : str
        Path to folder with image's content
    s2i_image : str
        Image to use as S2I template
    image_name : str
        Name of the final image
    s2i_runtime_image : str = None
        Optional runtime image
    """

    s2i_folder = request.param["s2i_folder"]
    s2i_image = request.param["s2i_image"]
    image_name = request.param["image_name"]
    s2i_runtime_image = request.param.get("s2i_runtime_image", None)

    image_name = create_s2i_image(s2i_folder, s2i_image, image_name, s2i_runtime_image)
    kind_load_image(image_name)

    return image_name


def do_s2i_python_version():
    return get_seldon_version()
