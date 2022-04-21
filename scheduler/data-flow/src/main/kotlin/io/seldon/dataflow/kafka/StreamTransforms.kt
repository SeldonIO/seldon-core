package io.seldon.dataflow.kafka

import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse
import org.apache.kafka.streams.kstream.KStream
import org.apache.kafka.streams.kstream.ValueTransformerSupplier

fun <T> KStream<T, TRecord>.filterForPipeline(pipelineName: String): KStream<T, TRecord> {
    return this
        .transformValues(ValueTransformerSupplier { HeaderFilter(pipelineName) })
        .filterNot { _, value -> value == null }
}

fun <T> KStream<T, ByteArray>.unmarshallInferenceV2(): KStream<T, ModelInferResponse> {
    return this
        .mapValues { bytes -> ModelInferResponse.parseFrom(bytes) }
}

fun <T> KStream<T, ModelInferRequest>.marshallInferenceV2(): KStream<T, ByteArray> {
    return this
        .mapValues { request -> request.toByteArray() }
}

fun <T> KStream<T, ModelInferResponse>.convertToRequests(
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
                        val convertedTensor = convertOutputToInputTensor(newName, tensor, response.rawOutputContentsCount>0)

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
