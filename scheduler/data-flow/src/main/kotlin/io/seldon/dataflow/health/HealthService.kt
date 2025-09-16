/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import io.klogging.noCoLogger
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope

/**
 * Service for managing and executing health checks
 */
class HealthService {
    private val logger = noCoLogger(HealthService::class)
    private val livenessChecks = mutableListOf<HealthCheck>()
    private val readinessChecks = mutableListOf<HealthCheck>()
    private val startupChecks = mutableListOf<HealthCheck>()

    /**
     * Register a health check for liveness probe
     */
    fun addLivenessCheck(healthCheck: HealthCheck) {
        livenessChecks.add(healthCheck)
        logger.debug("Added liveness health check: ${healthCheck.name}")
    }

    /**
     * Register a health check for readiness probe
     */
    fun addReadinessCheck(healthCheck: HealthCheck) {
        readinessChecks.add(healthCheck)
        logger.debug("Added readiness health check: ${healthCheck.name}")
    }

    /**
     * Register a health check for startup probe
     */
    fun addStartupCheck(healthCheck: HealthCheck) {
        startupChecks.add(healthCheck)
        logger.debug("Added startup health check: ${healthCheck.name}")
    }

    /**
     * Execute liveness health checks
     */
    suspend fun checkLiveness(): HealthCheckResult {
        return executeHealthChecks(livenessChecks, "liveness")
    }

    /**
     * Execute readiness health checks
     */
    suspend fun checkReadiness(): HealthCheckResult {
        return executeHealthChecks(readinessChecks, "readiness")
    }

    /**
     * Execute startup health checks
     */
    suspend fun checkStartup(): HealthCheckResult {
        return executeHealthChecks(startupChecks, "startup")
    }

    private suspend fun executeHealthChecks(
        checks: List<HealthCheck>,
        type: String,
    ): HealthCheckResult {
        return try {
            val results =
                coroutineScope {
                    checks.map { healthCheck ->
                        async {
                            try {
                                healthCheck.name to healthCheck.check()
                            } catch (e: Exception) {
                                logger.error("Health check ${healthCheck.name} failed with exception", e)
                                healthCheck.name to
                                    HealthStatus(
                                        isHealthy = false,
                                        message = "Health check failed: ${e.message}",
                                        details = mapOf("error" to e.javaClass.simpleName),
                                    )
                            }
                        }
                    }.awaitAll().toMap()
                }

            val overallHealthy = results.values.all { it.isHealthy }
            val status = if (overallHealthy) "UP" else "DOWN"

            logger.debug("$type health check result: $status (${results.size} checks)")

            HealthCheckResult(
                status = status,
                checks = results,
                overallHealthy = overallHealthy,
            )
        } catch (e: Exception) {
            logger.error("Failed to execute $type health checks", e)
            HealthCheckResult(
                status = "DOWN",
                checks =
                    mapOf(
                        "system" to
                            HealthStatus(
                                isHealthy = false,
                                message = "Failed to execute health checks: ${e.message}",
                                details = mapOf("error" to e.javaClass.simpleName),
                            ),
                    ),
                overallHealthy = false,
            )
        }
    }
}
