import pytest
from seldon_core.tf_helper import _TF_MISSING

skipif_tf_missing = pytest.mark.skipif(_TF_MISSING, reason="tensorflow is not present")
