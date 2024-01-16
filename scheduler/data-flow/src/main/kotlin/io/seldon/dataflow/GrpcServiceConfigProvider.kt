/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import io.grpc.Status

object GrpcServiceConfigProvider {
    // Details: https://github.com/grpc/proposal/blob/master/A6-client-retries.md#validation-of-retrypolicy
    // Example: https://github.com/grpc/grpc-java/blob/v1.35.0/examples/src/main/resources/io/grpc/examples/retrying/retrying_service_config.json
    // However does not work: https://github.com/grpc/grpc-kotlin/issues/277
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