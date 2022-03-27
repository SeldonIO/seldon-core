package io.seldon.mlops.inference.v2

import io.grpc.CallOptions
import io.grpc.CallOptions.DEFAULT
import io.grpc.Channel
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.ServerServiceDefinition
import io.grpc.ServerServiceDefinition.builder
import io.grpc.ServiceDescriptor
import io.grpc.Status
import io.grpc.Status.UNIMPLEMENTED
import io.grpc.StatusException
import io.grpc.kotlin.AbstractCoroutineServerImpl
import io.grpc.kotlin.AbstractCoroutineStub
import io.grpc.kotlin.ClientCalls
import io.grpc.kotlin.ClientCalls.unaryRpc
import io.grpc.kotlin.ServerCalls
import io.grpc.kotlin.ServerCalls.unaryServerMethodDefinition
import io.grpc.kotlin.StubFor
import io.seldon.mlops.inference.v2.GRPCInferenceServiceGrpc.getServiceDescriptor
import kotlin.String
import kotlin.coroutines.CoroutineContext
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.jvm.JvmOverloads
import kotlin.jvm.JvmStatic

/**
 * Holder for Kotlin coroutine-based client and server APIs for inference.GRPCInferenceService.
 */
object GRPCInferenceServiceGrpcKt {
  const val SERVICE_NAME: String = GRPCInferenceServiceGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = GRPCInferenceServiceGrpc.getServiceDescriptor()

  val serverLiveMethod: MethodDescriptor<V2Dataplane.ServerLiveRequest,
      V2Dataplane.ServerLiveResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getServerLiveMethod()

  val serverReadyMethod: MethodDescriptor<V2Dataplane.ServerReadyRequest,
      V2Dataplane.ServerReadyResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getServerReadyMethod()

  val modelReadyMethod: MethodDescriptor<V2Dataplane.ModelReadyRequest,
      V2Dataplane.ModelReadyResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelReadyMethod()

  val serverMetadataMethod: MethodDescriptor<V2Dataplane.ServerMetadataRequest,
      V2Dataplane.ServerMetadataResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getServerMetadataMethod()

  val modelMetadataMethod: MethodDescriptor<V2Dataplane.ModelMetadataRequest,
      V2Dataplane.ModelMetadataResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelMetadataMethod()

  val modelInferMethod: MethodDescriptor<V2Dataplane.ModelInferRequest,
      V2Dataplane.ModelInferResponse>
    @JvmStatic
    get() = GRPCInferenceServiceGrpc.getModelInferMethod()

  /**
   * A stub for issuing RPCs to a(n) inference.GRPCInferenceService service as suspending
   * coroutines.
   */
  @StubFor(GRPCInferenceServiceGrpc::class)
  class GRPCInferenceServiceCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<GRPCInferenceServiceCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions):
        GRPCInferenceServiceCoroutineStub = GRPCInferenceServiceCoroutineStub(channel, callOptions)

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun serverLive(request: V2Dataplane.ServerLiveRequest, headers: Metadata = Metadata()):
        V2Dataplane.ServerLiveResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getServerLiveMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun serverReady(request: V2Dataplane.ServerReadyRequest, headers: Metadata =
        Metadata()): V2Dataplane.ServerReadyResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getServerReadyMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun modelReady(request: V2Dataplane.ModelReadyRequest, headers: Metadata = Metadata()):
        V2Dataplane.ModelReadyResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelReadyMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun serverMetadata(request: V2Dataplane.ServerMetadataRequest, headers: Metadata =
        Metadata()): V2Dataplane.ServerMetadataResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getServerMetadataMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun modelMetadata(request: V2Dataplane.ModelMetadataRequest, headers: Metadata =
        Metadata()): V2Dataplane.ModelMetadataResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelMetadataMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun modelInfer(request: V2Dataplane.ModelInferRequest, headers: Metadata = Metadata()):
        V2Dataplane.ModelInferResponse = unaryRpc(
      channel,
      GRPCInferenceServiceGrpc.getModelInferMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the inference.GRPCInferenceService service based on Kotlin
   * coroutines.
   */
  abstract class GRPCInferenceServiceCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ServerLive.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun serverLive(request: V2Dataplane.ServerLiveRequest):
        V2Dataplane.ServerLiveResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ServerLive is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ServerReady.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun serverReady(request: V2Dataplane.ServerReadyRequest):
        V2Dataplane.ServerReadyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ServerReady is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ModelReady.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun modelReady(request: V2Dataplane.ModelReadyRequest):
        V2Dataplane.ModelReadyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelReady is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ServerMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun serverMetadata(request: V2Dataplane.ServerMetadataRequest):
        V2Dataplane.ServerMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ServerMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ModelMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun modelMetadata(request: V2Dataplane.ModelMetadataRequest):
        V2Dataplane.ModelMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for inference.GRPCInferenceService.ModelInfer.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun modelInfer(request: V2Dataplane.ModelInferRequest):
        V2Dataplane.ModelInferResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method inference.GRPCInferenceService.ModelInfer is unimplemented"))

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
    )).build()
  }
}
