package io.seldon.dataflow.kafka

import com.google.protobuf.kotlin.toByteString
import io.seldon.mlops.inference.v2.V2Dataplane
import java.nio.ByteBuffer
import java.nio.ByteOrder

enum class DataType {
    BOOL, UINT8, UINT16, UINT32, UINT64, INT8, INT16, INT32, INT64, FP16, FP32, FP64, BYTES
}

fun convertRequestToRawInputContents(request: V2Dataplane.ModelInferRequest): V2Dataplane.ModelInferRequest {
    val builder = request.toBuilder()
    request.inputsList.forEachIndexed { idx, input ->
        when (DataType.valueOf(input.datatype)) {
            DataType.UINT8 -> {
                val v = input.contents.uintContentsList.flatMap { ByteBuffer.allocate(1)
                    .order(ByteOrder.LITTLE_ENDIAN).put(it.toByte()).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.UINT16 -> {
                val v = input.contents.uintContentsList.flatMap { ByteBuffer.allocate(UShort.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putShort(it.toShort()).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.UINT32 -> {
                val v = input.contents.uintContentsList.flatMap { ByteBuffer.allocate(UInt.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putInt(it).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.UINT64 -> {
                val v = input.contents.uint64ContentsList.flatMap { ByteBuffer.allocate(ULong.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putLong(it).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.INT8 -> {
                val v = input.contents.intContentsList.flatMap { ByteBuffer.allocate(1)
                    .order(ByteOrder.LITTLE_ENDIAN).put(it.toByte()).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.INT16 -> {
                val v = input.contents.intContentsList.flatMap { ByteBuffer.allocate(Short.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putShort(it.toShort()).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.INT32 -> {
                val v = input.contents.intContentsList.flatMap { ByteBuffer.allocate(Int.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putInt(it).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.INT64 -> {
                val v = input.contents.int64ContentsList.flatMap { ByteBuffer.allocate(Long.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putLong(it).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.BOOL -> {
                val v = input.contents.boolContentsList.flatMap { ByteBuffer.allocate(1)
                    .order(ByteOrder.LITTLE_ENDIAN).put(if (it) {1} else {0}) .array().toList() }.toByteArray()
            }
            DataType.FP16, // Unclear if this is correct as will be stored as 4 byte floats
            DataType.FP32 -> {
                val v = input.contents.fp32ContentsList.flatMap { ByteBuffer.allocate(Float.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putFloat(it).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.FP64 -> {
                val v = input.contents.fp64ContentsList.flatMap { ByteBuffer.allocate(Double.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putDouble(it).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
            DataType.BYTES -> {
                val v = input.contents.bytesContentsList.flatMap { ByteBuffer.allocate(it.size() + 4)
                    .order(ByteOrder.LITTLE_ENDIAN)
                    .putInt(it.size())
                    .put(it.toByteArray()).array().toList() }.toByteArray()
                builder.addRawInputContents(v.toByteString())
            }
        }
        // Clear the contents now we have added the raw inputs
        builder.setInputs(idx, input.toBuilder().clearContents())
    }
    return builder.build()
}

fun convertResponseToRawOutputContents(request: V2Dataplane.ModelInferResponse): V2Dataplane.ModelInferResponse {
    val builder = request.toBuilder()
    request.outputsList.forEachIndexed { idx, output ->
        when (DataType.valueOf(output.datatype)) {
            DataType.UINT8 -> {
                val v = output.contents.uintContentsList.flatMap { ByteBuffer.allocate(1)
                    .order(ByteOrder.LITTLE_ENDIAN).put(it.toByte()).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.UINT16 -> {
                val v = output.contents.uintContentsList.flatMap { ByteBuffer.allocate(UShort.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putShort(it.toShort()).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.UINT32 -> {
                val v = output.contents.uintContentsList.flatMap { ByteBuffer.allocate(UInt.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putInt(it).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.UINT64 -> {
                val v = output.contents.uint64ContentsList.flatMap { ByteBuffer.allocate(ULong.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putLong(it).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.INT8 -> {
                val v = output.contents.intContentsList.flatMap { ByteBuffer.allocate(1)
                    .order(ByteOrder.LITTLE_ENDIAN).put(it.toByte()).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.INT16 -> {
                val v = output.contents.intContentsList.flatMap { ByteBuffer.allocate(Short.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putShort(it.toShort()).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.INT32 -> {
                val v = output.contents.intContentsList.flatMap { ByteBuffer.allocate(Int.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putInt(it).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.INT64 -> {
                val v = output.contents.int64ContentsList.flatMap { ByteBuffer.allocate(Long.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putLong(it).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.BOOL -> {
                val v = output.contents.boolContentsList.flatMap { ByteBuffer.allocate(1)
                    .order(ByteOrder.LITTLE_ENDIAN).put(if (it) {1} else {0}) .array().toList() }.toByteArray()
            }
            DataType.FP16, // Unclear if this is correct as will be stored as 4 byte floats
            DataType.FP32 -> {
                val v = output.contents.fp32ContentsList.flatMap { ByteBuffer.allocate(Float.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putFloat(it).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.FP64 -> {
                val v = output.contents.fp64ContentsList.flatMap { ByteBuffer.allocate(Double.SIZE_BYTES)
                    .order(ByteOrder.LITTLE_ENDIAN).putDouble(it).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
            DataType.BYTES -> {
                val v = output.contents.bytesContentsList.flatMap { ByteBuffer.allocate(it.size() + 4)
                    .order(ByteOrder.LITTLE_ENDIAN)
                    .putInt(it.size())
                    .put(it.toByteArray()).array().toList() }.toByteArray()
                builder.addRawOutputContents(v.toByteString())
            }
        }
        // Clear the contents now we have added the raw inputs
        builder.setOutputs(idx, output.toBuilder().clearContents())
    }
    return builder.build()
}

