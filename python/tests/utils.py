import pytest
import importlib


TF_MISSING = importlib.util.find_spec("tensorflow") is None

skipif_tf_missing = pytest.mark.skipif(
    TF_MISSING, reason="tensorflow is not present")
