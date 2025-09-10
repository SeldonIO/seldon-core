/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import io.klogging.noCoLogger
import io.seldon.dataflow.PipelineSubscriber
import kotlinx.coroutines.isActive

/**
 * Health check for internal service health - verifies the service is alive and can potentially recover from external issues
 */
class ServiceHealthCheck(
    private val pipelineSubscriber: PipelineSubscriber,
) : HealthCheck {
    override val name = "service"
    private val logger = noCoLogger(ServiceHealthCheck::class)

    override suspend fun check(): HealthStatus {
        return try {
            // Check if the main coroutine scope is still active
            val scopeActive = pipelineSubscriber.scope.isActive

            // Check if the dispatcher is still functional
            val dispatcherActive = pipelineSubscriber.dispatcher.isActive

            val isHealthy = scopeActive && dispatcherActive

            val message =
                when {
                    !scopeActive -> "Main coroutine scope is not active"
                    !dispatcherActive -> "Thread dispatcher is shutdown"
                    else -> "Service is running normally"
                }

            HealthStatus(
                isHealthy = isHealthy,
                message = message,
                details =
                    mapOf(
                        "scopeActive" to scopeActive.toString(),
                        "dispatcherActive" to dispatcherActive.toString(),
                    ),
            )
        } catch (e: Exception) {
            logger.error("Unexpected error in service health check", e)
            HealthStatus(
                isHealthy = false,
                message = "Service health check error: ${e.message}",
                details = mapOf("error" to e.javaClass.simpleName),
            )
        }
    }
}
