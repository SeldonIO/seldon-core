import flatbuffers


import seldon_core.fbs.SeldonMessage
import seldon_core.fbs.Data
import seldon_core.fbs.DefaultData
import seldon_core.fbs.Tensor
import seldon_core.fbs.SeldonRPC
import seldon_core.fbs.SeldonPayload
import seldon_core.fbs.Status
import seldon_core.fbs.StatusValue
import seldon_core.fbs.SeldonProtocolVersion
import seldon_core.fbs.SeldonMethod
import seldon_core.fbs as fbs
from seldon_core.seldon_flatbuffers import FlatbuffersInvalidMessage

def NumpyArrayToSeldonRPC(arr, names):
    builder = flatbuffers.Builder(32768)
    if len(names) > 0:
        str_offsets = []
        for i in range(len(names)):
            str_offsets.append(builder.CreateString(names[i]))
        fbs.DefaultData.DefaultDataStartNamesVector(builder, len(str_offsets))
        for i in reversed(range(len(str_offsets))):
            builder.PrependUOffsetTRelative(str_offsets[i])
        namesOffset = builder.EndVector(len(str_offsets))
    fbs.Tensor.TensorStartShapeVector(builder, len(arr.shape))
    for i in reversed(range(len(arr.shape))):
        builder.PrependInt32(arr.shape[i])
    sOffset = builder.EndVector(len(arr.shape))
    arr = arr.flatten()
    fbs.Tensor.TensorStartValuesVector(builder, len(arr))
    for i in reversed(range(len(arr))):
        builder.PrependFloat64(arr[i])
    vOffset = builder.EndVector(len(arr))
    fbs.Tensor.TensorStart(builder)
    fbs.Tensor.TensorAddShape(builder, sOffset)
    fbs.Tensor.TensorAddValues(builder, vOffset)
    tensor = fbs.Tensor.TensorEnd(builder)

    fbs.DefaultData.DefaultDataStart(builder)
    fbs.DefaultData.DefaultDataAddTensor(builder, tensor)
    fbs.DefaultData.DefaultDataAddNames(builder, namesOffset)
    defData = fbs.DefaultData.DefaultDataEnd(builder)

    fbs.Status.StatusStart(builder)
    fbs.Status.StatusAddCode(builder, 200)
    fbs.Status.StatusAddStatus(builder, fbs.StatusValue.StatusValue.SUCCESS)
    status = fbs.Status.StatusEnd(builder)

    fbs.SeldonMessage.SeldonMessageStart(builder)
    fbs.SeldonMessage.SeldonMessageAddProtocol(
        builder, fbs.SeldonProtocolVersion.SeldonProtocolVersion.V1)
    fbs.SeldonMessage.SeldonMessageAddStatus(builder, status)
    fbs.SeldonMessage.SeldonMessageAddDataType(
        builder, fbs.Data.Data.DefaultData)
    fbs.SeldonMessage.SeldonMessageAddData(builder, defData)
    seldonMessage = fbs.SeldonMessage.SeldonMessageEnd(builder)

    fbs.SeldonRPC.SeldonRPCStart(builder)
    fbs.SeldonRPC.SeldonRPCAddMethod(
        builder, fbs.SeldonMethod.SeldonMethod.PREDICT)
    fbs.SeldonRPC.SeldonRPCAddMessageType(
        builder, fbs.SeldonPayload.SeldonPayload.SeldonMessage)
    fbs.SeldonRPC.SeldonRPCAddMessage(builder, seldonMessage)
    seldonRPC = fbs.SeldonRPC.SeldonRPCEnd(builder)

    builder.FinishSizePrefixed(seldonRPC)
    return builder.Output()


def SeldonRPCToNumpyArray(data):
    seldon_msg = fbs.SeldonMessage.SeldonMessage.GetRootAsSeldonMessage(
        data, 0)
    if seldon_msg.Protocol() == fbs.SeldonProtocolVersion.SeldonProtocolVersion.V1:
        if seldon_msg.DataType() == fbs.Data.Data.DefaultData:
            defData = fbs.DefaultData.DefaultData()
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
            "Message does not have correct protocol: " + str(seldon_msg.Protocol()))
