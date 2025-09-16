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
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.Test
import strikt.api.expectThat
import strikt.assertions.isEqualTo
import strikt.assertions.isFalse
import strikt.assertions.isTrue
import java.util.concurrent.TimeUnit

class GrpcHealthCheckTest {
    @Test
    fun `should return healthy status when grpc channel is ready`() =
        runBlocking {
            // Given
            val mockChannel = createMockChannel(ConnectivityState.READY)
            val healthCheck = GrpcHealthCheck(mockChannel, "test-service")

            // When
            val result = healthCheck.check()

            // Then
            expectThat(result) {
                get { isHealthy }.isTrue()
                get { message }.isEqualTo("gRPC connection is ready")
                get { details["service"] }.isEqualTo("test-service")
                get { details["state"] }.isEqualTo("READY")
                get { details["target"] }.isEqualTo("test-service:9000")
            }
        }

    @Test
    fun `should return healthy status when grpc channel is connecting`() =
        runBlocking {
            // Given
            val mockChannel = createMockChannel(ConnectivityState.CONNECTING)
            val healthCheck = GrpcHealthCheck(mockChannel, "test-service")

            // When
            val result = healthCheck.check()

            // Then
            expectThat(result) {
                get { isHealthy }.isTrue()
                get { message }.isEqualTo("gRPC connection is establishing")
                get { details["state"] }.isEqualTo("CONNECTING")
            }
        }

    @Test
    fun `should return healthy status when grpc channel is idle`() =
        runBlocking {
            // Given
            val mockChannel = createMockChannel(ConnectivityState.IDLE)
            val healthCheck = GrpcHealthCheck(mockChannel, "test-service")

            // When
            val result = healthCheck.check()

            // Then
            expectThat(result) {
                get { isHealthy }.isTrue()
                get { message }.isEqualTo("gRPC connection is idle but available")
                get { details["state"] }.isEqualTo("IDLE")
            }
        }

    @Test
    fun `should return unhealthy status when grpc channel has transient failure`() =
        runBlocking {
            // Given
            val mockChannel = createMockChannel(ConnectivityState.TRANSIENT_FAILURE)
            val healthCheck = GrpcHealthCheck(mockChannel, "test-service")

            // When
            val result = healthCheck.check()

            // Then
            expectThat(result) {
                get { isHealthy }.isFalse()
                get { message }.isEqualTo("gRPC connection has transient failure")
                get { details["state"] }.isEqualTo("TRANSIENT_FAILURE")
            }
        }

    @Test
    fun `should return unhealthy status when grpc channel is shutdown`() =
        runBlocking {
            // Given
            val mockChannel = createMockChannel(ConnectivityState.SHUTDOWN)
            val healthCheck = GrpcHealthCheck(mockChannel, "test-service")

            // When
            val result = healthCheck.check()

            // Then
            expectThat(result) {
                get { isHealthy }.isFalse()
                get { message }.isEqualTo("gRPC connection is shutdown")
                get { details["state"] }.isEqualTo("SHUTDOWN")
            }
        }

    private fun createMockChannel(state: ConnectivityState): ManagedChannel {
        return object : ManagedChannel() {
            override fun getState(requestConnection: Boolean) = state

            override fun authority() = "test-service:9000"

            override fun shutdown() = this

            override fun shutdownNow() = this

            override fun isShutdown() = false

            override fun isTerminated() = false

            override fun awaitTermination(
                timeout: Long,
                unit: TimeUnit,
            ) = false

            override fun notifyWhenStateChanged(
                source: ConnectivityState,
                callback: Runnable,
            ) {}

            override fun <RequestT, ResponseT> newCall(
                methodDescriptor: io.grpc.MethodDescriptor<RequestT, ResponseT>,
                callOptions: io.grpc.CallOptions,
            ) = throw UnsupportedOperationException("Mock implementation")

            override fun resetConnectBackoff() {}
        }
    }
}
