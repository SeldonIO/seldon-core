import os
import importlib
import sys

# Variable to check if TF is present or not
_TF_MISSING = importlib.util.find_spec("tensorflow") is None

def _patch_tf_protos():
    """
    We make the `tensorflow` dependency optional.
    However, our protobuffer depends on some of TF's protobuffers.
    When `tensorflow` is not present, this methods patches that dependency
    inserting into the syspath the replacement package containing the TF's
    protobuffers we require.
    """
    seldon_core_path = os.path.dirname(__file__)
    tf_protos_path = os.path.join(seldon_core_path, "tensorflow")
    sys.path.insert(0, tf_protos_path)

