from tornado.tcpserver import TCPServer
from tornado.iostream import StreamClosedError
from tornado import gen
import tornado.ioloop

from fbs.SeldonMessage import *
from fbs.Data import *
from fbs.DefaultData import *
from fbs.Tensor import *
from fbs.SeldonRPC import *
from fbs.SeldonPayload import *

import sys
import numpy as np

def SeldonRPCToNumpyArray(data):
    seldon_rpc = SeldonRPC.GetRootAsSeldonRPC(data,0)
    if seldon_rpc.MessageType() == SeldonPayload.SeldonMessage:
        seldon_msg = SeldonMessage()
        seldon_msg.Init(seldon_rpc.Message().Bytes,seldon_rpc.Message().Pos)
        if seldon_msg.DataType() == Data.DefaultData:
            print("Its a default data")                    
            #data = seldon_msg.Data()
            defData = DefaultData()
            defData.Init(seldon_msg.Data().Bytes,seldon_msg.Data().Pos)
            tensor = defData.Tensor()
            shape = []
            for i in range(tensor.ShapeLength()):
                shape.append(tensor.Shape(i))
                print(shape)
            values = tensor.ValuesAsNumpy()
            return values
        else:
            return None
    else:
        return None

# Take a numpy array and create a SeldonRPC message
# Creates a local flat buffers builder
def NumpyArrayToSeldonRPC(arr):
    builder = flatbuffers.Builder(30000)
    TensorStartValuesVector(builder,len(arr))
    for i in reversed(range(len(arr))):
        builder.PrependFloat64(arr[i])
    vOffset = builder.EndVector(len(arr))
    TensorStartShapeVector(builder,len(arr.shape))
    for i in reversed(range(len(arr.shape))):
        builder.PrependInt32(arr.shape[i])
    sOffset = builder.EndVector(len(arr.shape))
    TensorStart(builder)
    TensorAddShape(builder,sOffset)
    TensorAddValues(builder,vOffset)
    tensor = TensorEnd(builder)
    DefaultDataStart(builder)
    DefaultDataAddTensor(builder,tensor)
    defData = DefaultDataEnd(builder)
    SeldonMessageStart(builder)
    SeldonMessageAddDataType(builder,Data.DefaultData)
    SeldonMessageAddData(builder,defData)
    seldonMessage = SeldonMessageEnd(builder)
    builder.FinishSizePrefixed(seldonMessage)
    return builder.Output()
