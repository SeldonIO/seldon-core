from seldon_core.imports_helper import _TF_MISSING

if not _TF_MISSING:
    # Let tensorflow shadow these imports if present
    from tensorflow.core.framework import tensor_pb2, types_pb2, tensor_shape_pb2
