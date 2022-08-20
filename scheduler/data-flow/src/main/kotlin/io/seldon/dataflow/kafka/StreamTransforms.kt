package io.seldon.dataflow.kafka

import io.seldon.dataflow.kafka.headers.PipelineNameFilter
import io.seldon.dataflow.kafka.headers.AlibiDetectRemover
import io.seldon.mlops.chainer.ChainerOuterClass.Batch
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse
import org.apache.kafka.streams.kstream.KStream
import org.apache.kafka.streams.kstream.ValueTransformerSupplier

fun <T> KStream<T, TRecord>.filterForPipeline(pipelineName: String): KStream<T, TRecord> {
    return this
        .transformValues(ValueTransformerSupplier { PipelineNameFilter(pipelineName) })
        .filterNot { _, value -> value == null }
}

fun <T> KStream<T, TRecord>.headerRemover(): KStream<T, TRecord> {
    return this
        .transformValues(ValueTransformerSupplier { AlibiDetectRemover() })
}

fun <T> KStream<T, ByteArray>.unmarshallInferenceV2Response(): KStream<T, ModelInferResponse> {
    return this
        .mapValues { bytes -> ModelInferResponse.parseFrom(bytes) }
}

fun <T> KStream<T, ModelInferRequest>.marshallInferenceV2Request(): KStream<T, ByteArray> {
    return this
        .mapValues { request -> request.toByteArray() }
}

fun <T> KStream<T, ByteArray>.unmarshallInferenceV2Request(): KStream<T, ModelInferRequest> {
    return this
        .mapValues { bytes -> ModelInferRequest.parseFrom(bytes) }
}

fun <T> KStream<T, ModelInferResponse>.marshallInferenceV2Response(): KStream<T, ByteArray> {
    return this
        .mapValues { request -> request.toByteArray() }
}


fun <T> KStream<T, ModelInferResponse>.convertToRequest(
    inputTopic: TopicName,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>
): KStream<T, ModelInferRequest> {
    return this
        .mapValues { inferResponse ->
            convertToRequest(
                inferResponse,
                desiredTensors,
                tensorRenaming,
                inputTopic,
            )
        }
}

/**
 * Batch model inference requests based on the strategy defined in the pipeline specification
 */
fun KStream<String, ModelInferRequest>.batchMessages(batchProperties: Batch): KStream<String, ModelInferRequest> {
    return when (batchProperties.size) {
        0 -> this
        else -> this
            .transform({ BatchProcessor(batchProperties.size) }, BatchProcessor.STATE_STORE_ID)
    }
}

/**
 * Convert the output from one model (a response) to the input for another model (a request).
 */
private fun convertToRequest(
    response: ModelInferResponse,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>,
    inputTopic: TopicName,
): ModelInferRequest {
    return ModelInferRequest
        .newBuilder()
        .setId(response.id)
        .putAllParameters(response.parametersMap)
        .apply {
            // Loop instead of `addAllInputs` to minimise intermediate memory usage, as tensors can be large
            response.outputsList
                .forEachIndexed { idx, tensor ->
                    if (tensor.name in desiredTensors || desiredTensors == null || desiredTensors.isEmpty()) {
                        val newName = tensorRenaming
                            .getOrDefault(
                                "${inputTopic}.${tensor.name}",
                                tensor.name,
                            )
                        val convertedTensor =
                            convertOutputToInputTensor(newName, tensor, response.rawOutputContentsCount > 0)

                        addInputs(convertedTensor)
                        if (idx < response.rawOutputContentsCount) {
                            // TODO - should add in appropriate index for raw input contents!
                            addRawInputContents(
                                response.getRawOutputContents(idx)
                            )
                        }
                    }
                }
        }.build()
}

private fun convertOutputToInputTensor(
    tensorName: TensorName,
    output: ModelInferResponse.InferOutputTensor,
    rawContents: Boolean
): ModelInferRequest.InferInputTensor {
    val req = ModelInferRequest.InferInputTensor
        .newBuilder()
        .setName(tensorName)
        .setDatatype(output.datatype)
        .addAllShape(output.shapeList)
        .putAllParameters(output.parametersMap)
    if (!rawContents) {
        req.setContents(output.contents)
    }
    return req.build()
}

fun <T> KStream<T, ModelInferRequest>.filterRequests(
    inputTopic: TopicName,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>
): KStream<T, ModelInferRequest> {
    return this
        .mapValues { inferResponse ->
            filterRequest(
                inferResponse,
                desiredTensors,
                tensorRenaming,
                inputTopic,
            )
        }
}

private fun filterRequest(
    request: ModelInferRequest,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>,
    inputTopic: TopicName,
): ModelInferRequest {
    return ModelInferRequest
        .newBuilder()
        .setId(request.id)
        .putAllParameters(request.parametersMap)
        .apply {
            // Loop instead of `addAllInputs` to minimise intermediate memory usage, as tensors can be large
            request.inputsList
                .forEachIndexed { idx, tensor ->
                    if (tensor.name in desiredTensors || desiredTensors == null || desiredTensors.isEmpty()) {
                        val newName = tensorRenaming
                            .getOrDefault(
                                "${inputTopic}.${tensor.name}",
                                tensor.name,
                            )

                        val convertedTensor = createInputTensor(newName, tensor, request.rawInputContentsCount>0)

                        addInputs(convertedTensor)
                        if (idx < request.rawInputContentsCount) {
                            // TODO - should add in appropriate index for raw input contents!
                            addRawInputContents(
                                request.getRawInputContents(idx)
                            )
                        }
                    }
                }
        }.build()
}

private fun createInputTensor(
    tensorName: TensorName,
    input: ModelInferRequest.InferInputTensor,
    rawContents: Boolean
): ModelInferRequest.InferInputTensor {
    val req = ModelInferRequest.InferInputTensor
        .newBuilder()
        .setName(tensorName)
        .setDatatype(input.datatype)
        .addAllShape(input.shapeList)
        .putAllParameters(input.parametersMap)
    if (!rawContents) {
        req.setContents(input.contents)
    }
    return req.build()
}


fun <T> KStream<T, ModelInferResponse>.filterResponses(
    inputTopic: TopicName,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>
): KStream<T, ModelInferResponse> {
    return this
        .mapValues { inferResponse ->
            filterResponse(
                inferResponse,
                desiredTensors,
                tensorRenaming,
                inputTopic,
            )
        }
}

private fun filterResponse(
    response: ModelInferResponse,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>,
    inputTopic: TopicName,
): ModelInferResponse {
    return ModelInferResponse
        .newBuilder()
        .setId(response.id)
        .putAllParameters(response.parametersMap)
        .apply {
            // Loop instead of `addAllInputs` to minimise intermediate memory usage, as tensors can be large
            response.outputsList
                .forEachIndexed { idx, tensor ->
                    if (tensor.name in desiredTensors || desiredTensors == null || desiredTensors.isEmpty()) {
                        val newName = tensorRenaming
                            .getOrDefault(
                                "${inputTopic}.${tensor.name}",
                                tensor.name,
                            )

                        val convertedTensor = createOutputTensor(newName, tensor, response.rawOutputContentsCount>0)

                        addOutputs(convertedTensor)
                        if (idx < response.rawOutputContentsCount) {
                            // TODO - should add in appropriate index for raw input contents!
                            addRawOutputContents(
                                response.getRawOutputContents(idx)
                            )
                        }
                    }
                }
        }.build()
}

private fun createOutputTensor(
    tensorName: TensorName,
    input: ModelInferResponse.InferOutputTensor,
    rawContents: Boolean
): ModelInferResponse.InferOutputTensor {
    val res = ModelInferResponse.InferOutputTensor
        .newBuilder()
        .setName(tensorName)
        .setDatatype(input.datatype)
        .addAllShape(input.shapeList)
        .putAllParameters(input.parametersMap)
    if (!rawContents) {
        res.setContents(input.contents)
    }
    return res.build()
}

fun <T> KStream<T, ModelInferRequest>.convertToResponse(
    inputTopic: TopicName,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>
): KStream<T, ModelInferResponse> {
    return this
        .mapValues { inferResponse ->
            convertToResponse(
                inferResponse,
                desiredTensors,
                tensorRenaming,
                inputTopic,
            )
        }
}

/**
 * Convert the output from one model (a response) to the input for another model (a request).
 */
private fun convertToResponse(
    request: ModelInferRequest,
    desiredTensors: Set<TensorName>?,
    tensorRenaming: Map<TensorName, TensorName>,
    inputTopic: TopicName,
): ModelInferResponse {
    return ModelInferResponse
        .newBuilder()
        .setId(request.id)
        .putAllParameters(request.parametersMap)
        .apply {
            // Loop instead of `addAllInputs` to minimise intermediate memory usage, as tensors can be large
            request.inputsList
                .forEachIndexed { idx, tensor ->
                    if (tensor.name in desiredTensors || desiredTensors == null || desiredTensors.isEmpty()) {
                        val newName = tensorRenaming
                            .getOrDefault(
                                "${inputTopic}.${tensor.name}",
                                tensor.name,
                            )
                        val convertedTensor = convertInputToOutputTensor(newName, tensor, request.rawInputContentsCount>0)

                        addOutputs(convertedTensor)
                        if (idx < request.rawInputContentsCount) {
                            // TODO - should add in appropriate index for raw input contents!
                            addRawOutputContents(
                                request.getRawInputContents(idx)
                            )
                        }
                    }
                }
        }.build()
}

private fun convertInputToOutputTensor(
    tensorName: TensorName,
    input: ModelInferRequest.InferInputTensor,
    rawContents: Boolean
): ModelInferResponse.InferOutputTensor {
    val req = ModelInferResponse.InferOutputTensor
        .newBuilder()
        .setName(tensorName)
        .setDatatype(input.datatype)
        .addAllShape(input.shapeList)
        .putAllParameters(input.parametersMap)
    if (!rawContents) {
        req.setContents(input.contents)
    }
    return req.build()
}