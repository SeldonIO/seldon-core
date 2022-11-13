/*
Copyright 2022 Seldon Technologies Ltd.

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