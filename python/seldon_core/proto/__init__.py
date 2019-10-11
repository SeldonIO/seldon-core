import logging
import textwrap
from seldon_core.tf_helper import _TF_MISSING, _patch_tf_protos

logger = logging.getLogger(__name__)

if _TF_MISSING:
    notice = textwrap.dedent("""
        Tensorflow is not installed.
        If you want to use `tftensor` and Tensorflow's data types
        install `tensorflow` or install `seldon_core` as

            $ pip install seldon_core[tensorflow]
    """)
    logger.info(notice)
    _patch_tf_protos()

