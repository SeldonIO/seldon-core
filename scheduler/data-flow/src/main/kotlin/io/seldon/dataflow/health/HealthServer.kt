/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import io.klogging.noCoLogger
import io.ktor.http.HttpStatusCode
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.application.install
import io.ktor.server.engine.EmbeddedServer
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import io.ktor.server.netty.NettyApplicationEngine
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import io.ktor.server.response.respond
import io.ktor.server.routing.get
import io.ktor.server.routing.routing
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json

/**
 * HTTP server for Kubernetes health probes
 */
class HealthServer(
    private val port: Int,
    private val healthService: HealthService,
    private val scope: CoroutineScope,
) {
    private val logger = noCoLogger(HealthServer::class)
    private var server: EmbeddedServer<NettyApplicationEngine, NettyApplicationEngine.Configuration>? = null
    private var serverJob: Job? = null

    /**
     * Start the health server
     */
    fun start() {
        if (server != null) {
            logger.warn("Health server is already running")
            return
        }

        logger.info("Starting health server on port $port")

        server =
            embeddedServer(Netty, port = port, host = "0.0.0.0", watchPaths = emptyList()) {
                install(ContentNegotiation) {
                    json(
                        Json {
                            prettyPrint = true
                            isLenient = true
                        },
                    )
                }

                routing {
                    get("/live") {
                        val result = healthService.checkLiveness()
                        val statusCode = if (result.overallHealthy) HttpStatusCode.OK else HttpStatusCode.ServiceUnavailable
                        if (statusCode != HttpStatusCode.OK) {
                            logger.warn("Live health check failed: ${result.status}")
                        } else {
                            logger.debug("Live health check passed: ${result.status}")
                        }
                        call.response.status(statusCode)
                        call.respond(result)
                    }

                    get("/ready") {
                        val result = healthService.checkReadiness()
                        val statusCode = if (result.overallHealthy) HttpStatusCode.OK else HttpStatusCode.ServiceUnavailable
                        if (statusCode != HttpStatusCode.OK) {
                            logger.warn("Ready health check failed: ${result.status}")
                        } else {
                            logger.debug("Ready health check passed: ${result.status}")
                        }
                        call.response.status(statusCode)
                        call.respond(result)
                    }

                    get("/startup") {
                        try {
                            val result = healthService.checkStartup()
                            val statusCode = if (result.overallHealthy) HttpStatusCode.OK else HttpStatusCode.ServiceUnavailable
                            if (statusCode != HttpStatusCode.OK) {
                                logger.warn("Startup health check failed: ${result.status}")
                            } else {
                                logger.debug("Startup health check passed: ${result.status}")
                            }
                            call.response.status(statusCode)
                            call.respond(result)
                        } catch (e: Exception) {
                            logger.error("Exception occurred during startup health check: ${e.message}", e)
                            val errorResult =
                                HealthCheckResult(
                                    overallHealthy = false,
                                    status = "DOWN",
                                    checks =
                                        mapOf(
                                            "startup-error" to
                                                HealthStatus(
                                                    isHealthy = false,
                                                    message = "Startup health check failed with exception: ${e.message}",
                                                    details = mapOf("exception" to e.javaClass.simpleName),
                                                ),
                                        ),
                                )
                            call.response.status(HttpStatusCode.InternalServerError)
                            call.respond(errorResult)
                        }
                    }
                }
            }

        serverJob =
            scope.launch {
                try {
                    server?.start(wait = false)
                } catch (e: Exception) {
                    logger.error("Health server failed", e)
                }
            }

        logger.info("Health server started successfully on port $port")
    }

    /**
     * Stop the health server
     */
    fun stop() {
        logger.info("Stopping health server")

        serverJob?.cancel()
        server?.stop(1000, 5000)

        server = null
        serverJob = null

        logger.info("Health server stopped")
    }
}
