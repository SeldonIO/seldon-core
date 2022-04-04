package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.inference.v2.V2Dataplane
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.kstream.*
import org.apache.kafka.streams.kstream.internals.JoinedInternal
import java.time.Duration

/**
 * A *transformer* which joins multiple input streams into a single output stream.
 */
class Joiner(
    properties: KafkaProperties,
    internal val inputTopics: Set<TopicName>,
    internal val outputTopic: TopicName,
    internal val tensors: Map<TopicName, Set<TensorName>>?,
    internal val pipelineName: String,
    internal val tensorMap: Map<TensorName, TensorName>,
) : Transformer {

    private val streams: KafkaStreams by lazy {
        val producerSerde: Produced<RequestId, TRecord> = Produced.with(Serdes.String(), Serdes.ByteArray())
        val builder = StreamsBuilder()

        buildStream(builder, inputTopics)
            .filterNot { _, value -> value == null } // Not sure if this is needed
            .to(outputTopic, producerSerde)

        KafkaStreams(builder.build(), properties)
    }

    private fun buildStream(builder: StreamsBuilder, inputTopics: Set<TopicName>): KStream<RequestId,ByteArray> {
        val consumerSerde: Consumed<RequestId, TRecord> = Consumed.with(Serdes.String(), Serdes.ByteArray())
        val joinedSerde: StreamJoined<RequestId, TRecord, TRecord> = StreamJoined.with(Serdes.String(), Serdes.ByteArray(), Serdes.ByteArray())
        val inputTopic = inputTopics.first()
        val restTopics = inputTopics.minus(inputTopic)
        val left = builder
            .stream(inputTopic, consumerSerde)
            .transformValues(::getHeaderFilter)
            .filterNot { _, value -> value == null }
            .mapValues { bytes -> V2Dataplane.ModelInferResponse.parseFrom(bytes) }
            .mapValues{ v -> convertResponseToRequest(inputTopic, tensors?.get(inputTopic), v)}
            .mapValues { message -> message.toByteArray() }
        if (inputTopics.size > 1) {
            val right = buildStream(builder, restTopics)
            val leftJoined = left.join(right,::requestJoiner, JoinWindows.of(Duration.ofMinutes(5)),
                joinedSerde
           )
            return leftJoined
        } else {
            return left
        }
    }

    private fun requestJoiner(left: ByteArray, right: ByteArray): ByteArray {
        val leftDe =  V2Dataplane.ModelInferRequest.parseFrom(left)
        val rightDe = V2Dataplane.ModelInferRequest.parseFrom(right)
        val res = V2Dataplane.ModelInferRequest
            .newBuilder()
            .setId(leftDe.id)
            .putAllParameters(leftDe.parametersMap)
            .addAllInputs(leftDe.inputsList)
            .addAllInputs(rightDe.inputsList)
            .addAllRawInputContents(leftDe.rawInputContentsList)
            .addAllRawInputContents(rightDe.rawInputContentsList)
            .build()
        logger.info("Joined request ${res}")
        return res.toByteArray()
    }

    private fun getHeaderFilter() = HeaderFilter(pipelineName)

    /**
     * Convert the output from one model (a response) to the input for another model (a request).
     */
    private fun convertResponseToRequest(inputTopic: TopicName, tensors: Set<TensorName>?, response: V2Dataplane.ModelInferResponse): V2Dataplane.ModelInferRequest {
        return V2Dataplane.ModelInferRequest
            .newBuilder()
            .setId(response.id)
            .putAllParameters(response.parametersMap)
            .apply {
                // Loop instead of `addAllInputs` to minimise intermediate memory usage, as tensors can be large
                response.outputsList
                    //.filter { tensor ->
                    //    tensor.name in tensors
                   // }
                    .forEachIndexed { idx, tensor ->
                        logger.info("Checking for ${tensor.name} in ${tensors} for topic ${inputTopic}")
                        if (tensor.name in tensors) {
                            logger.info("Adding tensor ${tensor.name}")
                            addInputs(
                                convertOutputToInputTensor( tensorMap.getOrDefault(inputTopic+"."+tensor.name, tensor.name), tensor)
                            )
                            addRawInputContents(response.getRawOutputContents(idx))
                        }

                    }
            }
            .build()
    }


    private fun convertOutputToInputTensor(
        tensorName: TensorName,
        output: V2Dataplane.ModelInferResponse.InferOutputTensor,
    ): V2Dataplane.ModelInferRequest.InferInputTensor {
        return V2Dataplane.ModelInferRequest.InferInputTensor
            .newBuilder()
            .setName(tensorName)
            .setDatatype(output.datatype)
            .addAllShape(output.shapeList)
            .putAllParameters(output.parametersMap)
            .setContents(output.contents)
            .build()
    }

    override fun start() {
        streams.cleanUp() // TODO - make configurable via CLI
        Joiner.logger.info("starting for ($inputTopics) -> ($outputTopic)")
        streams.start()
    }

    override fun stop() {
        logger.info("Stopping joiner")
        streams.close()
    }

    companion object {
        private val logger = noCoLogger(Joiner::class)
    }
}