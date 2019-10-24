import pytest
from k8s_utils import *
from s2i_utils import *


@pytest.fixture(scope="module")
def s2i_python_version():
    return do_s2i_python_version()


#### Implementations below


def do_s2i_python_version():
    return get_s2i_python_version()
