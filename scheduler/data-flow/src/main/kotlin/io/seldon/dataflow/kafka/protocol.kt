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
import io.seldon.mlops.inference.v2.V2Dataplane
import java.nio.ByteBuffer
import java.nio.ByteOrder

enum class DataType {
    BOOL, UINT8, UINT16, UINT32, UINT64, INT8, INT16, INT32, INT64, FP16, FP32, FP64, BYTES
}

fun convertRequestToRawInputContents(request: V2Dataplane.ModelInferRequest): V2Dataplane.ModelInferRequest {
    val builder = request.toBuilder()
    request.inputsList.forEachIndexed { idx, input ->
        val v = when (DataType.valueOf(input.datatype)) {
            DataType.UINT8 -> {
                input.contents.uintContentsList.flatMap {
                    ByteBuffer
                        .allocate(1)
                        .put(it.toByte())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.UINT16 -> {
                input.contents.uintContentsList.flatMap {
                    ByteBuffer
                        .allocate(UShort.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putShort(it.toShort())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.UINT32 -> {
                input.contents.uintContentsList.flatMap {
                    ByteBuffer
                        .allocate(UInt.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putInt(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.UINT64 -> {
                input.contents.uint64ContentsList.flatMap {
                    ByteBuffer
                        .allocate(ULong.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putLong(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT8 -> {
                input.contents.intContentsList.flatMap {
                    ByteBuffer
                        .allocate(1)
                        .put(it.toByte())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT16 -> {
                input.contents.intContentsList.flatMap {
                    ByteBuffer
                        .allocate(Short.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putShort(it.toShort())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT32 -> {
                input.contents.intContentsList.flatMap {
                    ByteBuffer
                        .allocate(Int.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putInt(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT64 -> {
                input.contents.int64ContentsList.flatMap {
                    ByteBuffer
                        .allocate(Long.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putLong(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.BOOL -> {
                input.contents.boolContentsList.flatMap {
                    ByteBuffer
                        .allocate(1)
                        .put(if (it) {1} else {0})
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.FP16, // may need to handle this separately in future
            DataType.FP32 -> {
                input.contents.fp32ContentsList.flatMap {
                    ByteBuffer
                        .allocate(Float.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putFloat(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.FP64 -> {
                input.contents.fp64ContentsList.flatMap {
                    ByteBuffer
                        .allocate(Double.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putDouble(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.BYTES -> {
                input.contents.bytesContentsList.flatMap {
                    ByteBuffer
                        .allocate(it.size() + Int.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putInt(it.size())
                        .put(it.toByteArray())
                        .array()
                        .toList()
                }.toByteArray()
            }
        }
        // Add raw contents
        builder.addRawInputContents(v.toByteString())
        // Clear the contents now we have added the raw inputs
        builder.getInputsBuilder(idx).clearContents()
    }
    return builder.build()
}

fun convertResponseToRawOutputContents(request: V2Dataplane.ModelInferResponse): V2Dataplane.ModelInferResponse {
    val builder = request.toBuilder()
    request.outputsList.forEachIndexed { idx, output ->
        val v = when (DataType.valueOf(output.datatype)) {
            DataType.UINT8 -> {
                output.contents.uintContentsList.flatMap {
                    ByteBuffer
                        .allocate(1)
                        .put(it.toByte())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.UINT16 -> {
                output.contents.uintContentsList.flatMap {
                    ByteBuffer
                        .allocate(UShort.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putShort(it.toShort())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.UINT32 -> {
                output.contents.uintContentsList.flatMap {
                    ByteBuffer
                        .allocate(UInt.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putInt(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.UINT64 -> {
                output.contents.uint64ContentsList.flatMap {
                    ByteBuffer
                        .allocate(ULong.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putLong(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT8 -> {
                output.contents.intContentsList.flatMap {
                    ByteBuffer
                        .allocate(1)
                        .put(it.toByte())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT16 -> {
                output.contents.intContentsList.flatMap {
                    ByteBuffer
                        .allocate(Short.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putShort(it.toShort())
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT32 -> {
                output.contents.intContentsList.flatMap {
                    ByteBuffer
                        .allocate(Int.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putInt(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.INT64 -> {
                output.contents.int64ContentsList.flatMap {
                    ByteBuffer
                        .allocate(Long.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putLong(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.BOOL -> {
                output.contents.boolContentsList.flatMap {
                    ByteBuffer.allocate(1)
                        .put(if (it) {1} else {0})
                        .array().toList()
                }.toByteArray()
            }
            DataType.FP16, // may need to handle this separately in future
            DataType.FP32 -> {
                output.contents.fp32ContentsList.flatMap {
                    ByteBuffer
                        .allocate(Float.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putFloat(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.FP64 -> {
                output.contents.fp64ContentsList.flatMap {
                    ByteBuffer
                        .allocate(Double.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putDouble(it)
                        .array()
                        .toList()
                }.toByteArray()
            }
            DataType.BYTES -> {
                output.contents.bytesContentsList.flatMap {
                    ByteBuffer
                        .allocate(it.size() + Int.SIZE_BYTES)
                        .order(ByteOrder.LITTLE_ENDIAN)
                        .putInt(it.size())
                        .put(it.toByteArray())
                        .array()
                        .toList()
                }.toByteArray()
            }
        }
        // Add raw contents
        builder.addRawOutputContents(v.toByteString())
        // Clear the contents now we have added the raw outputs
        builder.getOutputsBuilder(idx).clearContents()
    }
    return builder.build()
}

