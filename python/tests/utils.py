import pytest
from seldon_core.imports_helper import _TF_MISSING, _GCS_MISSING

skipif_tf_missing = pytest.mark.skipif(_TF_MISSING, reason="tensorflow is not present")
skipif_gcs_missing = pytest.mark.skipif(
    _GCS_MISSING, reason="google-cloud-storage is not present"
)
