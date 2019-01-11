from tornado.tcpserver import TCPServer
from tornado.iostream import StreamClosedError
from tornado import gen
import tornado.ioloop

from flatbuffers.number_types import (
    UOffsetTFlags, SOffsetTFlags, VOffsetTFlags)

import sys
import numpy as np

from seldon_core.fbs.SeldonMessage import *
from seldon_core.fbs.Data import *
from seldon_core.fbs.DefaultData import *
from seldon_core.fbs.Tensor import *
from seldon_core.fbs.SeldonRPC import *
from seldon_core.fbs.SeldonPayload import *
from seldon_core.fbs.Status import *
from seldon_core.fbs.StatusValue import *
from seldon_core.fbs.SeldonProtocolVersion import *
from seldon_core.fbs.SeldonRPC import *


class FlatbuffersInvalidMessage(Exception):
    def __init__(self, msg=None):
        super(FlatbuffersInvalidMessage, self).__init__(msg)


def SeldonRPCToNumpyArray(data):
    seldon_rpc = SeldonRPC.GetRootAsSeldonRPC(data, 0)
    if seldon_rpc.MessageType() == SeldonPayload.SeldonMessage:
        seldon_msg = SeldonMessage()
        seldon_msg.Init(seldon_rpc.Message().Bytes, seldon_rpc.Message().Pos)
        if seldon_msg.Protocol() == SeldonProtocolVersion.V1:
            if seldon_msg.DataType() == Data.DefaultData:
                defData = DefaultData()
                defData.Init(seldon_msg.Data().Bytes, seldon_msg.Data().Pos)
                names = []
                for i in range(defData.NamesLength()):
                    names.append(defData.Names(i))
                tensor = defData.Tensor()
                shape = []
                for i in range(tensor.ShapeLength()):
                    shape.append(tensor.Shape(i))
                values = tensor.ValuesAsNumpy()
                values = values.reshape(shape)
                return (values, names)
            else:
                raise FlatbuffersInvalidMessage(
                    "Message is not of type DefaultData")
        else:
            raise FlatbuffersInvalidMessage(
                "Message does not have correct protocol: " + str(seldon_rpc.Protocol()))
    else:
        raise FlatbuffersInvalidMessage("Message is not a SeldonMessage")


def CreateErrorMsg(msg):
    builder = flatbuffers.Builder(4096)

    msg_offset = builder.CreateString(msg)

    StatusStart(builder)
    StatusAddCode(builder, 500)
    StatusAddInfo(builder, msg_offset)
    StatusAddStatus(builder, StatusValue.FAILURE)
    status = StatusEnd(builder)

    SeldonMessageStart(builder)
    SeldonMessageAddStatus(builder, status)
    seldonMessage = SeldonMessageEnd(builder)
    builder.FinishSizePrefixed(seldonMessage)
    return builder.Output()


# Take a numpy array and create a SeldonRPC message
# Creates a local flat buffers builder
def NumpyArrayToSeldonRPC(arr, names):
    builder = flatbuffers.Builder(32768)
    if len(names) > 0:
        str_offsets = []
        for i in range(len(names)):
            str_offsets.append(builder.CreateString(names[i]))
        DefaultDataStartNamesVector(builder, len(str_offsets))
        for i in reversed(range(len(str_offsets))):
            builder.PrependUOffsetTRelative(str_offsets[i])
        namesOffset = builder.EndVector(len(str_offsets))
    TensorStartShapeVector(builder, len(arr.shape))
    for i in reversed(range(len(arr.shape))):
        builder.PrependInt32(arr.shape[i])
    sOffset = builder.EndVector(len(arr.shape))
    arr = arr.flatten()

    # TensorStartValuesVector(builder,len(arr))
    # for i in reversed(range(len(arr))):
    #    builder.PrependFloat64(arr[i])
    #vOffset = builder.EndVector(len(arr))
    vOffset = CreateNumpyVector(builder, arr)

    TensorStart(builder)
    TensorAddShape(builder, sOffset)
    TensorAddValues(builder, vOffset)
    tensor = TensorEnd(builder)

    DefaultDataStart(builder)
    DefaultDataAddTensor(builder, tensor)
    if len(names) > 0:
        DefaultDataAddNames(builder, namesOffset)
    defData = DefaultDataEnd(builder)

    StatusStart(builder)
    StatusAddCode(builder, 200)
    StatusAddStatus(builder, StatusValue.SUCCESS)
    status = StatusEnd(builder)

    SeldonMessageStart(builder)
    SeldonMessageAddProtocol(builder, SeldonProtocolVersion.V1)
    SeldonMessageAddStatus(builder, status)
    SeldonMessageAddDataType(builder, Data.DefaultData)
    SeldonMessageAddData(builder, defData)
    seldonMessage = SeldonMessageEnd(builder)

    builder.FinishSizePrefixed(seldonMessage)
    return builder.Output()


def CreateNumpyVector(builder, x):
    """CreateNumpyVector writes a numpy array into the buffer."""

    if np is None:
        # Numpy is required for this feature
        raise NumpyRequiredForThisFeature("Numpy was not found.")

    if not isinstance(x, np.ndarray):
        raise TypeError("non-numpy-ndarray passed to CreateNumpyVector")

    if x.ndim > 1:
        raise TypeError("multidimensional-ndarray passed to CreateNumpyVector")

    builder.StartVector(x.itemsize, x.size, x.dtype.alignment)

    # Ensure little endian byte ordering
    if x.dtype.str[0] == "<":
        x_lend = x
    else:
        x_lend = x.byteswap(inplace=False)

    # Calculate total length
    l = UOffsetTFlags.py_type(x_lend.itemsize * x_lend.size)
    # @cond FLATBUFFERS_INTERNAL
    builder.head = UOffsetTFlags.py_type(builder.Head() - l)
    # @endcond

    # tobytes ensures c_contiguous ordering
    builder.Bytes[builder.Head():builder.Head() +
                  l] = x_lend.tobytes(order='C')

    return builder.EndVector(x.size)
