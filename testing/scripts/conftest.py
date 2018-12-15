import pytest
from k8s_utils import *
from s2i_utils import *

@pytest.fixture(scope="module")
def single_namespace_seldon_helm(request):
    create_seldon_single_namespace_helm(request)
    port_forward(request)

@pytest.fixture(scope="module")
def clusterwide_seldon_helm(request):
    create_seldon_clusterwide_helm(request)
    port_forward(request)

@pytest.fixture(scope="module")
def single_namespace_seldon_ksonnet(request):
    create_seldon_single_namespace_ksonnet(request)
    port_forward(request)

@pytest.fixture(scope="module")
def clusterwide_seldon_ksonnet(request):
    create_seldon_clusterwide_ksonnet(request)
    port_forward(request)

@pytest.fixture(scope="module")
def setup_python_s2i(request):
    build_python_s2i_images()
    

@pytest.fixture(scope="module")
def s2i_python_version():
    return get_s2i_python_version()

@pytest.fixture(scope="module")
def setup_local_docker_repo(request):
    create_docker_repo(request)
    port_forward_docker_repo(request)

