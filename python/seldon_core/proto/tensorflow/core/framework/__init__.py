import logging
import textwrap

logger = logging.getLogger(__name__)

try:
    # Let tensorflow shadow these imports if present
    from . import (
        tensor_pb2, types_pb2, tensor_shape_pb2)
except ImportError:
    notice = textwrap.dedent("""
        Tensorflow is not installed.
        If you want to use `tftensor` and Tensorflow's data types
        install `tensorflow` or install `seldon_core` as

            $ pip install seldon_core[tensorflow]
    """)
    logger.notice(notice)
