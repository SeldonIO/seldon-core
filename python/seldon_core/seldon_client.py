
import tensorflow as tf
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import numpy as np
import grpc

def grpc_request_external(data,endpoint="localhost:5000"):
    datadef = prediction_pb2.DefaultData(
            tftensor=tf.make_tensor_proto(data)
        )

    request = prediction_pb2.SeldonMessage(data = datadef)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    response = stub.Predict(request=request)
    return response



def grpc_request_internal(data,endpoint="localhost:5000"):
    datadef = prediction_pb2.DefaultData(
            tftensor=tf.make_tensor_proto(data)
        )

    request = prediction_pb2.SeldonMessage(data = datadef)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.ModelStub(channel)
    response = stub.Predict(request=request)
    return response


arr = np.array([1,2,3])
#response = grpc_request_internal(arr)
response = grpc_request_external(arr,"0.0.0.0:10000")
print(response)

