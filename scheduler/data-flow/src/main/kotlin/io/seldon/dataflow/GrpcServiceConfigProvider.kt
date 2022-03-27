package io.seldon.dataflow

import io.grpc.Status

object GrpcServiceConfigProvider {
    // Details: https://github.com/grpc/proposal/blob/master/A6-client-retries.md#validation-of-retrypolicy
    // Example: https://github.com/grpc/grpc-java/blob/v1.35.0/examples/src/main/resources/io/grpc/examples/retrying/retrying_service_config.json
    val config = mapOf<String, Any>(
        "methodConfig" to listOf(
            mapOf(
                "name" to listOf(
                    mapOf(
                        "service" to "io.seldon.mlops.chainer.Chainer",
                        "method" to "SubscribePipelineUpdates",
                    ),
                ),
                "retryPolicy" to mapOf(
                    "maxAttempts" to "100",
                    "initialBackoff" to "1s",
                    "maxBackoff" to "30s",
                    "backoffMultiplier" to 1.5,
                    "retryableStatusCodes" to listOf(
                        Status.UNAVAILABLE.code.toString(),
                        Status.CANCELLED.code.toString(),
                        Status.FAILED_PRECONDITION.code.toString(),
                    )
                )
            ),
        ),
    )
}