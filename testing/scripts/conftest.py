import pytest
from k8s_utils import *
from s2i_utils import *
from java_utils import *
from go_utils import *

@pytest.fixture(scope="module")
def clusterwide_seldon_helm(request):
    version = get_seldon_version()
    create_seldon_clusterwide_helm(request,version)
    port_forward(request)

@pytest.fixture(scope="module")
def setup_python_s2i(request):
    build_python_s2i_images()

@pytest.fixture(scope="module")
def s2i_python_version():
    return get_s2i_python_version()

@pytest.fixture(scope="session")
def seldon_images(request):
    create_docker_repo(request)
    port_forward_docker_repo(request)
    build_java_images()
    build_go_images()

@pytest.fixture(scope="session")
def seldon_version():
    return get_seldon_version()
