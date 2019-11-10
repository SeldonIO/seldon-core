import pytest
from seldon_e2e_utils import get_s2i_python_version


@pytest.fixture(scope="module")
def s2i_python_version():
    return do_s2i_python_version()


#### Implementations below


def do_s2i_python_version():
    return get_s2i_python_version()
