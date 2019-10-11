import logging
import textwrap

logger = logging.getLogger(__name__)

try:
    from tensorflow.core.framework import tensor_pb2
except ImportError:
    notice = textwrap.dedent("""
        Tensorflow is not installed.
        If you want to use `tftensor` and Tensorflow's data types
        install `tensorflow` or install `seldon_core` as

            $ pip install seldon_core[tensorflow]
    """)
    logger.info(notice)
    from seldon_core.tf_proto._fallback import (
        tensor_pb2, resource_handle_pb2, tensor_shape_pb2, types_pb2)

