import pytest
from k8s_utils import *
from s2i_utils import *
from java_utils import *
from go_utils import *

@pytest.fixture(scope="module")
def clusterwide_seldon_helm(request):
    do_clusterwide_seldon_helm(request)

@pytest.fixture(scope="module")
def setup_python_s2i():
    do_setup_python_s2i()

@pytest.fixture(scope="module")
def s2i_python_version():
    return do_s2i_python_version()

@pytest.fixture(scope="session")
def seldon_images(request):
    do_seldon_images(request)

@pytest.fixture(scope="session")
def seldon_version():
    return get_seldon_version()

#### Implementatiosn below

def do_s2i_python_version():
    return get_s2i_python_version()

def do_clusterwide_seldon_helm(request=None):
    version = get_seldon_version()
    create_seldon_clusterwide_helm(request,version)
    if not request is None:
        port_forward(request)

def do_create_docker_repo(request=None):
    create_docker_repo(request)

def do_seldon_images(request=None):
    create_docker_repo(request)
    port_forward_docker_repo(request)
    build_java_images()
    version = get_seldon_version()
    build_go_images(version)

def do_setup_python_s2i():
    build_python_s2i_images()