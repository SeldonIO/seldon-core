/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.mlops.inference.v2

import io.grpc.CallOptions
import io.grpc.CallOptions.DEFAULT
import io.grpc.Channel
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.ServerServiceDefinition
import io.grpc.ServerServiceDefinition.builder
import io.grpc.ServiceDescriptor
import io.grpc.Status.UNIMPLEMENTED
import io.grpc.StatusException
import io.grpc.kotlin.AbstractCoroutineServerImpl
import io.grpc.kotlin.AbstractCoroutineStub
import io.grpc.kotlin.ClientCalls.bidiStreamingRpc
import io.grpc.kotlin.ClientCalls.unaryRpc
import io.grpc.kotlin.ServerCalls.bidiStreamingServerMethodDefinition
import io.grpc.kotlin.ServerCalls.unaryServerMethodDefinition
import io.grpc.kotlin.StubFor
import io.seldon.mlops.inference.v2.GRPCInferenceServiceGrpc.getServiceDescriptor
import kotlin.String
import kotlin.coroutines.CoroutineContext
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.jvm.JvmOverloads
import kotlin.jvm.JvmStatic
import kotlinx.coroutines.flow.Flow

/**
 * Holder for Kotlin coroutine-based client and server APIs for inference.GRPCInferenceService.
 */
public object GRPCInferenceServiceGrpcKt {
  public const val SERVICE_NAME: String = GRPCInferenceServiceGrpc.SERVICE_NAME

  @JvmStatic
  public val serviceDescriptor: ServiceDescriptor
    get() = getServiceDescriptor()

  public val serverLiveMethod:
      MethodDescriptor<V2Dataplane.ServerLiveRequest, V2Dataplane.ServerLiveResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getServerLiveMethod()

  public val serverReadyMethod:
      MethodDescriptor<V2Dataplane.ServerReadyRequest, V2Dataplane.ServerReadyResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getServerReadyMethod()

  public val modelReadyMethod:
      MethodDescriptor<V2Dataplane.ModelReadyRequest, V2Dataplane.ModelReadyResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelReadyMethod()

  public val serverMetadataMethod:
      MethodDescriptor<V2Dataplane.ServerMetadataRequest, V2Dataplane.ServerMetadataResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getServerMetadataMethod()

  public val modelMetadataMethod:
      MethodDescriptor<V2Dataplane.ModelMetadataRequest, V2Dataplane.ModelMetadataResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelMetadataMethod()

  public val modelInferMethod:
      MethodDescriptor<V2Dataplane.ModelInferRequest, V2Dataplane.ModelInferResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelInferMethod()

  public val modelStreamInferMethod:
      MethodDescriptor<V2Dataplane.ModelInferRequest, V2Dataplane.ModelInferResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelStreamInferMethod()

  public val repositoryIndexMethod:
      MethodDescriptor<V2Dataplane.RepositoryIndexRequest, V2Dataplane.RepositoryIndexResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getRepositoryIndexMethod()

  public val repositoryModelLoadMethod:
      MethodDescriptor<V2Dataplane.RepositoryModelLoadRequest, V2Dataplane.RepositoryModelLoadResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getRepositoryModelLoadMethod()

  public val repositoryModelUnloadMethod:
      MethodDescriptor<V2Dataplane.RepositoryModelUnloadRequest, V2Dataplane.RepositoryModelUnloadResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getRepositoryModelUnloadMethod()

  /**
   * A stub for issuing RPCs to a(n) inference.GRPCInferenceService service as suspending
   * coroutines.
   */
  @StubFor(GRPCInferenceServiceGrpc::class)
  public class GRPCInferenceServiceCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT,
  ) : AbstractCoroutineStub<GRPCInferenceServiceCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions):
        GRPCInferenceServiceCoroutineStub = GRPCInferenceServiceCoroutineStub(channel, callOptions)

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun serverLive(request: V2Dataplane.ServerLiveRequest, headers: Metadata =
        Metadata()): V2Dataplane.ServerLiveResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getServerLiveMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun serverReady(request: V2Dataplane.ServerReadyRequest, headers: Metadata =
        Metadata()): V2Dataplane.ServerReadyResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getServerReadyMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun modelReady(request: V2Dataplane.ModelReadyRequest, headers: Metadata =
        Metadata()): V2Dataplane.ModelReadyResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelReadyMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun serverMetadata(request: V2Dataplane.ServerMetadataRequest, headers: Metadata
        = Metadata()): V2Dataplane.ServerMetadataResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getServerMetadataMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun modelMetadata(request: V2Dataplane.ModelMetadataRequest, headers: Metadata =
        Metadata()): V2Dataplane.ModelMetadataResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelMetadataMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun modelInfer(request: V2Dataplane.ModelInferRequest, headers: Metadata =
        Metadata()): V2Dataplane.ModelInferResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelInferMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Returns a [Flow] that, when collected, executes this RPC and emits responses from the
     * server as they arrive.  That flow finishes normally if the server closes its response with
     * [`Status.OK`][io.grpc.Status], and fails by throwing a [StatusException] otherwise.  If
     * collecting the flow downstream fails exceptionally (including via cancellation), the RPC
     * is cancelled with that exception as a cause.
     *
     * The [Flow] of requests is collected once each time the [Flow] of responses is
     * collected. If collection of the [Flow] of responses completes normally or
     * exceptionally before collection of `requests` completes, the collection of
     * `requests` is cancelled.  If the collection of `requests` completes
     * exceptionally for any other reason, then the collection of the [Flow] of responses
     * completes exceptionally for the same reason and the RPC is cancelled with that reason.
     *
     * @param requests A [Flow] of request messages.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return A flow that, when collected, emits the responses from the server.
     */
    public fun modelStreamInfer(requests: Flow<V2Dataplane.ModelInferRequest>, headers: Metadata =
        Metadata()): Flow<V2Dataplane.ModelInferResponse> = bidiStreamingRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelStreamInferMethod(),
      requests,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun repositoryIndex(request: V2Dataplane.RepositoryIndexRequest,
        headers: Metadata = Metadata()): V2Dataplane.RepositoryIndexResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getRepositoryIndexMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun repositoryModelLoad(request: V2Dataplane.RepositoryModelLoadRequest,
        headers: Metadata = Metadata()): V2Dataplane.RepositoryModelLoadResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getRepositoryModelLoadMethod(),
      request,
      callOptions,
      headers
    )

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    public suspend fun repositoryModelUnload(request: V2Dataplane.RepositoryModelUnloadRequest,
        headers: Metadata = Metadata()): V2Dataplane.RepositoryModelUnloadResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getRepositoryModelUnloadMethod(),
      request,
      callOptions,
      headers
    )
  }

  /**
   * Skeletal implementation of the inference.GRPCInferenceService service based on Kotlin
   * coroutines.
   */
  public abstract class GRPCInferenceServiceCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext,
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ServerLive.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun serverLive(request: V2Dataplane.ServerLiveRequest):
        V2Dataplane.ServerLiveResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ServerLive is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ServerReady.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun serverReady(request: V2Dataplane.ServerReadyRequest):
        V2Dataplane.ServerReadyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ServerReady is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ModelReady.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun modelReady(request: V2Dataplane.ModelReadyRequest):
        V2Dataplane.ModelReadyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelReady is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ServerMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun serverMetadata(request: V2Dataplane.ServerMetadataRequest):
        V2Dataplane.ServerMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ServerMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ModelMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun modelMetadata(request: V2Dataplane.ModelMetadataRequest):
        V2Dataplane.ModelMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ModelInfer.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun modelInfer(request: V2Dataplane.ModelInferRequest):
        V2Dataplane.ModelInferResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelInfer is unimplemented"))

    /**
     * Returns a [Flow] of responses to an RPC for inference.GRPCInferenceService.ModelStreamInfer.
     *
     * If creating or collecting the returned flow fails with a [StatusException], the RPC
     * will fail with the corresponding [io.grpc.Status].  If it fails with a
     * [java.util.concurrent.CancellationException], the RPC will fail with status
     * `Status.CANCELLED`.  If creating
     * or collecting the returned flow fails for any other reason, the RPC will fail with
     * `Status.UNKNOWN` with the exception as a cause.
     *
     * @param requests A [Flow] of requests from the client.  This flow can be
     *        collected only once and throws [java.lang.IllegalStateException] on attempts to
     * collect
     *        it more than once.
     */
    public open fun modelStreamInfer(requests: Flow<V2Dataplane.ModelInferRequest>):
        Flow<V2Dataplane.ModelInferResponse> = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelStreamInfer is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.RepositoryIndex.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun repositoryIndex(request: V2Dataplane.RepositoryIndexRequest):
        V2Dataplane.RepositoryIndexResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.RepositoryIndex is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.RepositoryModelLoad.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend fun repositoryModelLoad(request: V2Dataplane.RepositoryModelLoadRequest):
        V2Dataplane.RepositoryModelLoadResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.RepositoryModelLoad is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.RepositoryModelUnload.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    public open suspend
        fun repositoryModelUnload(request: V2Dataplane.RepositoryModelUnloadRequest):
        V2Dataplane.RepositoryModelUnloadResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.RepositoryModelUnload is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getServerLiveMethod(),
      implementation = ::serverLive
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getServerReadyMethod(),
      implementation = ::serverReady
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getModelReadyMethod(),
      implementation = ::modelReady
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getServerMetadataMethod(),
      implementation = ::serverMetadata
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getModelMetadataMethod(),
      implementation = ::modelMetadata
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getModelInferMethod(),
      implementation = ::modelInfer
    ))
      .addMethod(bidiStreamingServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getModelStreamInferMethod(),
      implementation = ::modelStreamInfer
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getRepositoryIndexMethod(),
      implementation = ::repositoryIndex
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getRepositoryModelLoadMethod(),
      implementation = ::repositoryModelLoad
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = GRPCInferenceServiceGrpc.getRepositoryModelUnloadMethod(),
      implementation = ::repositoryModelUnload
    )).build()
  }
}
