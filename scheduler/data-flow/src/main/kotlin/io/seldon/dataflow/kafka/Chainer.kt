package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Produced

/**
 * A *transformer* for a single input stream to a single output stream.
 */
class Chainer(
    properties: KafkaProperties,
    internal val inputTopic: TopicName,
    internal val outputTopic: TopicName,
    internal val tensors: Set<TensorName>?,
    internal val pipelineName: String,
) : Transformer {
    private val streams: KafkaStreams by lazy {
        val consumerSerde: Consumed<RequestId, TRecord> = Consumed.with(Serdes.String(), Serdes.ByteArray())
        val producerSerde: Produced<RequestId, TRecord> = Produced.with(Serdes.String(), Serdes.ByteArray())
        val builder = StreamsBuilder()

        builder
            .stream(inputTopic, consumerSerde)
            .transformValues(::getHeaderFilter)
            .filterNot { _, value -> value == null }
            // TODO - allow parsing from JSON or Protobuf
            .mapValues { bytes -> ModelInferResponse.parseFrom(bytes) }
            .mapValues(::convertResponseToRequest)
            .mapValues { message -> message.toByteArray() }
            .to(outputTopic, producerSerde)
        // TODO - when does K-Streams send an ack?  On consuming or only once a new value has been produced?
        // TODO - wait until streams exists, if it does not already

        KafkaStreams(builder.build(), properties)
    }

    private fun getHeaderFilter() = HeaderFilter(pipelineName)

    /**
     * Convert the output from one model (a response) to the input for another model (a request).
     */
    private fun convertResponseToRequest(response: ModelInferResponse): ModelInferRequest {
        return ModelInferRequest
            .newBuilder()
            .setId(response.id)
            .putAllParameters(response.parametersMap)
            .apply {
                // Loop instead of `addAllInputs` to minimise intermediate memory usage, as tensors can be large
                response.outputsList
                    .filter { tensor ->
                        tensor.name in tensors
                    }
                    .forEach { tensor ->
                        addInputs(
                            convertOutputToInputTensor(tensor)
                        )
                    }
            }
            .build()
    }

    private fun convertOutputToInputTensor(
        output: ModelInferResponse.InferOutputTensor
    ): ModelInferRequest.InferInputTensor {
        return ModelInferRequest.InferInputTensor
            .newBuilder()
            .setName(output.name)
            .setDatatype(output.datatype)
            .addAllShape(output.shapeList)
            .putAllParameters(output.parametersMap)
            .setContents(output.contents)
            .build()
    }

    override fun start() {
        streams.cleanUp()
        logger.info("starting for ($inputTopic) -> ($outputTopic)")
        streams.start()
    }

    override fun stop() {
        streams.close()
    }

    companion object {
        private val logger = noCoLogger(Chainer::class)
    }
}