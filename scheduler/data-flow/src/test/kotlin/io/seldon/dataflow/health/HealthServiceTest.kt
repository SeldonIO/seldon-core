/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import kotlinx.coroutines.delay
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.Test
import strikt.api.expectThat
import strikt.assertions.contains
import strikt.assertions.isEqualTo
import strikt.assertions.isFalse
import strikt.assertions.isLessThan
import strikt.assertions.isNotNull
import strikt.assertions.isTrue

class HealthServiceTest {
    @Test
    fun `should return healthy status when all liveness checks pass`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val healthyCheck1 = createMockHealthCheck("check1", true, "All good")
            val healthyCheck2 = createMockHealthCheck("check2", true, "Working fine")

            healthService.addLivenessCheck(healthyCheck1)
            healthService.addLivenessCheck(healthyCheck2)

            // When
            val result = healthService.checkLiveness()

            // Then
            expectThat(result) {
                get { status }.isEqualTo("UP")
                get { overallHealthy }.isTrue()
                get { checks.size }.isEqualTo(2)
                get { checks["check1"]?.isHealthy }.isTrue()
                get { checks["check2"]?.isHealthy }.isTrue()
            }
        }

    @Test
    fun `should return unhealthy status when any liveness check fails`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val healthyCheck = createMockHealthCheck("healthy", true, "All good")
            val unhealthyCheck = createMockHealthCheck("unhealthy", false, "Something wrong")

            healthService.addLivenessCheck(healthyCheck)
            healthService.addLivenessCheck(unhealthyCheck)

            // When
            val result = healthService.checkLiveness()

            // Then
            expectThat(result) {
                get { status }.isEqualTo("DOWN")
                get { overallHealthy }.isFalse()
                get { checks.size }.isEqualTo(2)
                get { checks["healthy"]?.isHealthy }.isTrue()
                get { checks["unhealthy"]?.isHealthy }.isFalse()
                get { checks["unhealthy"]?.message }.isEqualTo("Something wrong")
            }
        }

    @Test
    fun `should return healthy status when all readiness checks pass`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val readyCheck1 = createMockHealthCheck("ready1", true, "Ready to serve")
            val readyCheck2 = createMockHealthCheck("ready2", true, "All systems go")

            healthService.addReadinessCheck(readyCheck1)
            healthService.addReadinessCheck(readyCheck2)

            // When
            val result = healthService.checkReadiness()

            // Then
            expectThat(result) {
                get { status }.isEqualTo("UP")
                get { overallHealthy }.isTrue()
                get { checks.size }.isEqualTo(2)
                get { checks["ready1"]?.isHealthy }.isTrue()
                get { checks["ready2"]?.isHealthy }.isTrue()
            }
        }

    @Test
    fun `should return unhealthy status when any readiness check fails`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val readyCheck = createMockHealthCheck("ready", true, "Ready to serve")
            val notReadyCheck = createMockHealthCheck("not-ready", false, "Still initializing")

            healthService.addReadinessCheck(readyCheck)
            healthService.addReadinessCheck(notReadyCheck)

            // When
            val result = healthService.checkReadiness()

            // Then
            expectThat(result) {
                get { status }.isEqualTo("DOWN")
                get { overallHealthy }.isFalse()
                get { checks.size }.isEqualTo(2)
                get { checks["ready"]?.isHealthy }.isTrue()
                get { checks["not-ready"]?.isHealthy }.isFalse()
                get { checks["not-ready"]?.message }.isEqualTo("Still initializing")
            }
        }

    @Test
    fun `should handle separate liveness and readiness checks`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val livenessCheck = createMockHealthCheck("liveness", true, "Liveness check")
            val readinessCheck = createMockHealthCheck("readiness", true, "Readiness check")

            healthService.addLivenessCheck(livenessCheck)
            healthService.addReadinessCheck(readinessCheck)

            // When
            val livenessResult = healthService.checkLiveness()
            val readinessResult = healthService.checkReadiness()

            // Then
            expectThat(livenessResult) {
                get { checks.size }.isEqualTo(1)
                get { checks["liveness"]?.isHealthy }.isTrue()
            }

            expectThat(readinessResult) {
                get { checks.size }.isEqualTo(1)
                get { checks["readiness"]?.isHealthy }.isTrue()
            }
        }

    @Test
    fun `should handle exceptions in health checks gracefully`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val failingCheck =
                object : HealthCheck {
                    override val name = "failing-check"

                    override suspend fun check(): HealthStatus {
                        throw RuntimeException("Unexpected error")
                    }
                }

            healthService.addLivenessCheck(failingCheck)

            // When
            val result = healthService.checkLiveness()

            // Then
            expectThat(result) {
                get { status }.isEqualTo("DOWN")
                get { overallHealthy }.isFalse()
                get { checks.size }.isEqualTo(1)
                get { checks["failing-check"]?.isHealthy }.isFalse()
                get { checks["failing-check"]?.message }.isNotNull().contains("Health check failed")
                get { checks["failing-check"]?.details?.get("error") }.isEqualTo("RuntimeException")
            }
        }

    @Test
    fun `should execute health checks in parallel`() =
        runBlocking {
            // Given
            val healthService = HealthService()
            val slowCheck1 =
                object : HealthCheck {
                    override val name = "slow1"

                    override suspend fun check(): HealthStatus {
                        delay(200)
                        return HealthStatus(true, "Slow check 1")
                    }
                }

            val slowCheck2 =
                object : HealthCheck {
                    override val name = "slow2"

                    override suspend fun check(): HealthStatus {
                        delay(200)
                        return HealthStatus(true, "Slow check 2")
                    }
                }

            healthService.addLivenessCheck(slowCheck1)
            healthService.addLivenessCheck(slowCheck2)

            // When
            val startTime = System.currentTimeMillis()
            val result = healthService.checkLiveness()
            val duration = System.currentTimeMillis() - startTime

            // Then
            expectThat(result) {
                get { status }.isEqualTo("UP")
                get { overallHealthy }.isTrue()
                get { checks.size }.isEqualTo(2)
            }

            // Should complete in less than 350ms (parallel execution)
            // If sequential, it would take at least 400ms
            expectThat(duration).isLessThan(350)
        }

    private fun createMockHealthCheck(
        name: String,
        isHealthy: Boolean,
        message: String,
    ): HealthCheck {
        return object : HealthCheck {
            override val name = name

            override suspend fun check(): HealthStatus {
                return HealthStatus(isHealthy, message)
            }
        }
    }
}
