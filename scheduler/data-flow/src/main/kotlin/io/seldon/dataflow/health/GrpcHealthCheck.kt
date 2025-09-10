/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import io.grpc.ConnectivityState
import io.grpc.ManagedChannel
import io.klogging.noCoLogger
import kotlinx.coroutines.withTimeoutOrNull

/**
 * Health check for gRPC channel connectivity
 */
class GrpcHealthCheck(
    private val channel: ManagedChannel,
    private val serviceName: String,
    private val timeoutMs: Long = 5000,
) : HealthCheck {
    override val name = "grpc-$serviceName"
    private val logger = noCoLogger(GrpcHealthCheck::class)

    override suspend fun check(): HealthStatus {
        return try {
            val result =
                withTimeoutOrNull(timeoutMs) {
                    val state = channel.getState(true) // request connection if idle

                    when (state) {
                        ConnectivityState.READY -> {
                            HealthStatus(
                                isHealthy = true,
                                message = "gRPC connection is ready",
                                details =
                                    mapOf(
                                        "service" to serviceName,
                                        "state" to state.name,
                                        "target" to channel.authority(),
                                    ),
                            )
                        }
                        ConnectivityState.CONNECTING -> {
                            HealthStatus(
                                isHealthy = true,
                                message = "gRPC connection is establishing",
                                details =
                                    mapOf(
                                        "service" to serviceName,
                                        "state" to state.name,
                                        "target" to channel.authority(),
                                    ),
                            )
                        }
                        ConnectivityState.IDLE -> {
                            HealthStatus(
                                isHealthy = true,
                                message = "gRPC connection is idle but available",
                                details =
                                    mapOf(
                                        "service" to serviceName,
                                        "state" to state.name,
                                        "target" to channel.authority(),
                                    ),
                            )
                        }
                        ConnectivityState.TRANSIENT_FAILURE -> {
                            HealthStatus(
                                isHealthy = false,
                                message = "gRPC connection has transient failure",
                                details =
                                    mapOf(
                                        "service" to serviceName,
                                        "state" to state.name,
                                        "target" to channel.authority(),
                                    ),
                            )
                        }
                        ConnectivityState.SHUTDOWN -> {
                            HealthStatus(
                                isHealthy = false,
                                message = "gRPC connection is shutdown",
                                details =
                                    mapOf(
                                        "service" to serviceName,
                                        "state" to state.name,
                                        "target" to channel.authority(),
                                    ),
                            )
                        }
                        else -> {
                            HealthStatus(
                                isHealthy = false,
                                message = "gRPC connection state unknown: $state",
                                details =
                                    mapOf(
                                        "service" to serviceName,
                                        "state" to state.name,
                                        "target" to channel.authority(),
                                    ),
                            )
                        }
                    }
                }

            result ?: HealthStatus(
                isHealthy = false,
                message = "gRPC health check timed out",
                details =
                    mapOf(
                        "service" to serviceName,
                        "timeout" to "${timeoutMs}ms",
                        "target" to channel.authority(),
                    ),
            )
        } catch (e: Exception) {
            logger.error("Unexpected error in gRPC health check for $serviceName", e)
            HealthStatus(
                isHealthy = false,
                message = "gRPC health check error: ${e.message}",
                details =
                    mapOf(
                        "service" to serviceName,
                        "error" to e.javaClass.simpleName,
                        "target" to channel.authority(),
                    ),
            )
        }
    }
}
