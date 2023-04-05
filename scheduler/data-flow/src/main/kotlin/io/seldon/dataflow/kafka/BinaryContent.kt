/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package io.seldon.dataflow.kafka

import com.google.protobuf.kotlin.toByteString
import io.seldon.mlops.inference.v2.V2Dataplane.InferTensorContents
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse
import java.nio.ByteBuffer
import java.nio.ByteOrder

enum class DataType {
    BOOL,
    BYTES,
    UINT8, UINT16, UINT32, UINT64,
    INT8, INT16, INT32, INT64,
    FP16, FP32, FP64,
}

fun ModelInferRequest.withBinaryContents(): ModelInferRequest {
    return this.toBuilder().run {
        inputsList.forEachIndexed { idx, input ->
            val v = when (DataType.valueOf(input.datatype)) {
                DataType.UINT8 -> input.contents.toUint8Bytes()
                DataType.UINT16 -> input.contents.toUint16Bytes()
                DataType.UINT32 -> input.contents.toUint32Bytes()
                DataType.UINT64 -> input.contents.toUint64Bytes()
                DataType.INT8 -> input.contents.toInt8Bytes()
                DataType.INT16 -> input.contents.toInt16Bytes()
                DataType.INT32 -> input.contents.toInt32Bytes()
                DataType.INT64 -> input.contents.toInt64Bytes()
                DataType.BOOL -> input.contents.toBoolBytes()
                DataType.FP16, // may need to handle this separately in future
                DataType.FP32 -> input.contents.toFp32Bytes()
                DataType.FP64 -> input.contents.toFp64Bytes()
                DataType.BYTES -> input.contents.toRawBytes()
            }

            // Add binary data and clear corresponding contents.
            addRawInputContents(v.toByteString())
            getInputsBuilder(idx).clearContents()
        }

        build()
    }
}

fun ModelInferResponse.withBinaryContents(): ModelInferResponse {
    return this.toBuilder().run {
        outputsList.forEachIndexed { idx, output ->
            val v = when (DataType.valueOf(output.datatype)) {
                DataType.UINT8 -> output.contents.toUint8Bytes()
                DataType.UINT16 -> output.contents.toUint16Bytes()
                DataType.UINT32 -> output.contents.toUint32Bytes()
                DataType.UINT64 -> output.contents.toUint64Bytes()
                DataType.INT8 -> output.contents.toInt8Bytes()
                DataType.INT16 -> output.contents.toInt16Bytes()
                DataType.INT32 -> output.contents.toInt32Bytes()
                DataType.INT64 -> output.contents.toInt64Bytes()
                DataType.BOOL -> output.contents.toBoolBytes()
                DataType.FP16, // may need to handle this separately in future
                DataType.FP32 -> output.contents.toFp32Bytes()
                DataType.FP64 -> output.contents.toFp64Bytes()
                DataType.BYTES -> output.contents.toRawBytes()
            }

            // Add binary data and clear corresponding contents.
            addRawOutputContents(v.toByteString())
            getOutputsBuilder(idx).clearContents()
        }

        build()
    }
}

fun InferTensorContents.toUint8Bytes(): ByteArray = this.uintContentsList
    .flatMap {
        ByteBuffer
            .allocate(1)
            .put(it.toByte())
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toUint16Bytes(): ByteArray = this.uintContentsList
    .flatMap {
        ByteBuffer
            .allocate(UShort.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putShort(it.toShort())
            .array()
            .toList()
    }.toByteArray()

fun InferTensorContents.toUint32Bytes(): ByteArray = this.uintContentsList
    .flatMap {
        ByteBuffer
            .allocate(UInt.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putInt(it)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toUint64Bytes(): ByteArray = this.uint64ContentsList
    .flatMap {
        ByteBuffer
            .allocate(ULong.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putLong(it)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toInt8Bytes(): ByteArray = this.intContentsList
    .flatMap {
        ByteBuffer
            .allocate(1)
            .put(it.toByte())
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toInt16Bytes(): ByteArray = this.intContentsList
    .flatMap {
        ByteBuffer
            .allocate(Short.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putShort(it.toShort())
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toInt32Bytes(): ByteArray = this.intContentsList
    .flatMap {
        ByteBuffer
            .allocate(Int.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putInt(it)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toInt64Bytes(): ByteArray = this.int64ContentsList
    .flatMap {
        ByteBuffer
            .allocate(Long.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putLong(it)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toFp32Bytes(): ByteArray = this.fp32ContentsList
    .flatMap {
        ByteBuffer
            .allocate(Float.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putFloat(it)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toFp64Bytes(): ByteArray = this.fp64ContentsList
    .flatMap {
        ByteBuffer
            .allocate(Double.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putDouble(it)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toBoolBytes(): ByteArray = this.boolContentsList
    .flatMap {
        ByteBuffer
            .allocate(1)
            .put(if (it) 1 else 0)
            .array()
            .toList()
    }
    .toByteArray()

fun InferTensorContents.toRawBytes(): ByteArray = this.bytesContentsList
    .flatMap {
        ByteBuffer
            .allocate(it.size() + Int.SIZE_BYTES)
            .order(ByteOrder.LITTLE_ENDIAN)
            .putInt(it.size())
            .put(it.toByteArray())
            .array()
            .toList()
    }
    .toByteArray()