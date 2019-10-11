import logging
from seldon_core.tf_helper import _TF_MISSING, _patch_tf_protos


if _TF_MISSING:
    logging.info("""
        Tensorflow is not installed.
        If you want to use `tftensor` and Tensorflow's data types
        install `tensorflow` or install `seldon_core` as

            $ pip install seldon_core[tensorflow]
    """)
    _patch_tf_protos()

