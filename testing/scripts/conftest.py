import pytest
from k8s_utils import *

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
