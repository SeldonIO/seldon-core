/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Test
import strikt.api.expectThat
import strikt.assertions.contains
import strikt.assertions.isEqualTo
import java.net.ServerSocket
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse

class HealthServerTest {
    private var healthServer: HealthServer? = null
    private val testScope = CoroutineScope(SupervisorJob())

    @AfterEach
    fun cleanup() {
        healthServer?.stop()
        healthServer = null
    }

    @Test
    fun `should start and stop health server successfully`() {
        // Given
        val healthService = HealthService()
        healthServer = HealthServer(0, healthService, testScope) // Port 0 for random port

        // When & Then - Should not throw
        healthServer!!.start()
        healthServer!!.stop()
    }

    @Test
    fun `should respond to health endpoints when started`() =
        runBlocking {
            // Given
            val port = findAvailablePort()
            val healthService = HealthService()
            val healthCheck = createMockHealthCheck("test", true, "All good")
            healthService.addLivenessCheck(healthCheck)
            healthService.addReadinessCheck(healthCheck)

            healthServer = HealthServer(port, healthService, testScope)
            healthServer!!.start()

            // Wait for server to start
            delay(500)

            val client = HttpClient.newHttpClient()

            // When & Then - Test liveness endpoint
            val livenessRequest =
                HttpRequest.newBuilder()
                    .uri(URI.create("http://localhost:$port/live"))
                    .GET()
                    .build()

            val livenessResponse = client.send(livenessRequest, HttpResponse.BodyHandlers.ofString())
            expectThat(livenessResponse.statusCode()).isEqualTo(200)
            expectThat(livenessResponse.body()).contains("UP")

            // Test readiness endpoint
            val readinessRequest =
                HttpRequest.newBuilder()
                    .uri(URI.create("http://localhost:$port/ready"))
                    .GET()
                    .build()

            val readinessResponse = client.send(readinessRequest, HttpResponse.BodyHandlers.ofString())
            expectThat(readinessResponse.statusCode()).isEqualTo(200)
            expectThat(readinessResponse.body()).contains("UP")

            // Test startup endpoint
            val startupRequest =
                HttpRequest.newBuilder()
                    .uri(URI.create("http://localhost:$port/startup"))
                    .GET()
                    .build()

            val startupResponse = client.send(startupRequest, HttpResponse.BodyHandlers.ofString())
            expectThat(startupResponse.statusCode()).isEqualTo(200)

            // Test root endpoint
            val rootRequest =
                HttpRequest.newBuilder()
                    .uri(URI.create("http://localhost:$port/"))
                    .GET()
                    .build()

            val rootResponse = client.send(rootRequest, HttpResponse.BodyHandlers.ofString())
            expectThat(rootResponse.statusCode()).isEqualTo(200)
            expectThat(rootResponse.body()).contains("Seldon DataFlow Engine Health Server")
        }

    @Test
    fun `should return 503 status when health checks fail`() =
        runBlocking {
            // Given
            val port = findAvailablePort()
            val healthService = HealthService()
            val unhealthyCheck = createMockHealthCheck("test", false, "Something wrong")
            healthService.addLivenessCheck(unhealthyCheck)
            healthService.addReadinessCheck(unhealthyCheck)

            healthServer = HealthServer(port, healthService, testScope)
            healthServer!!.start()

            // Wait for server to start
            delay(500)

            val client = HttpClient.newHttpClient()

            // When
            val request =
                HttpRequest.newBuilder()
                    .uri(URI.create("http://localhost:$port/live"))
                    .GET()
                    .build()

            val response = client.send(request, HttpResponse.BodyHandlers.ofString())

            // Then
            expectThat(response.statusCode()).isEqualTo(503) // Service Unavailable
            expectThat(response.body()).contains("DOWN")
            expectThat(response.body()).contains("Something wrong")
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

    private fun findAvailablePort(): Int {
        return ServerSocket(0).use { socket ->
            socket.localPort
        }
    }
}
