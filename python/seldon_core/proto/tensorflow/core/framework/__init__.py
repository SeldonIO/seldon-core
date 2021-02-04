from seldon_core.imports_helper import _TF_PRESENT

if _TF_PRESENT:
    # Let tensorflow shadow these imports if present
    from tensorflow.core.framework import tensor_pb2, tensor_shape_pb2, types_pb2
