package io.seldon.mlops.chainer

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
import io.grpc.kotlin.ClientCalls.serverStreamingRpc
import io.grpc.kotlin.ClientCalls.unaryRpc
import io.grpc.kotlin.ServerCalls.serverStreamingServerMethodDefinition
import io.grpc.kotlin.ServerCalls.unaryServerMethodDefinition
import io.grpc.kotlin.StubFor
import io.seldon.mlops.chainer.ChainerGrpc.getServiceDescriptor
import kotlin.String
import kotlin.coroutines.CoroutineContext
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.jvm.JvmOverloads
import kotlin.jvm.JvmStatic
import kotlinx.coroutines.flow.Flow

/**
 * Holder for Kotlin coroutine-based client and server APIs for seldon.mlops.chainer.Chainer.
 */
object ChainerGrpcKt {
  const val SERVICE_NAME: String = ChainerGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = ChainerGrpc.getServiceDescriptor()

  val subscribePipelineUpdatesMethod:
      MethodDescriptor<ChainerOuterClass.PipelineSubscriptionRequest,
      ChainerOuterClass.PipelineUpdateMessage>
    @JvmStatic
    get() = ChainerGrpc.getSubscribePipelineUpdatesMethod()

  val pipelineUpdateEventMethod: MethodDescriptor<ChainerOuterClass.PipelineUpdateStatusMessage,
      ChainerOuterClass.PipelineUpdateStatusResponse>
    @JvmStatic
    get() = ChainerGrpc.getPipelineUpdateEventMethod()

  /**
   * A stub for issuing RPCs to a(n) seldon.mlops.chainer.Chainer service as suspending coroutines.
   */
  @StubFor(ChainerGrpc::class)
  class ChainerCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<ChainerCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions): ChainerCoroutineStub =
        ChainerCoroutineStub(channel, callOptions)

    /**
     * Returns a [Flow] that, when collected, executes this RPC and emits responses from the
     * server as they arrive.  That flow finishes normally if the server closes its response with
     * [`Status.OK`][Status], and fails by throwing a [StatusException] otherwise.  If
     * collecting the flow downstream fails exceptionally (including via cancellation), the RPC
     * is cancelled with that exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return A flow that, when collected, emits the responses from the server.
     */
    fun subscribePipelineUpdates(request: ChainerOuterClass.PipelineSubscriptionRequest,
        headers: Metadata = Metadata()): Flow<ChainerOuterClass.PipelineUpdateMessage> =
        serverStreamingRpc(
      channel,
      ChainerGrpc.getSubscribePipelineUpdatesMethod(),
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
    suspend fun pipelineUpdateEvent(request: ChainerOuterClass.PipelineUpdateStatusMessage,
        headers: Metadata = Metadata()): ChainerOuterClass.PipelineUpdateStatusResponse = unaryRpc(
      channel,
      ChainerGrpc.getPipelineUpdateEventMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the seldon.mlops.chainer.Chainer service based on Kotlin coroutines.
   */
  abstract class ChainerCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns a [Flow] of responses to an RPC for
     * seldon.mlops.chainer.Chainer.SubscribePipelineUpdates.
     *
     * If creating or collecting the returned flow fails with a [StatusException], the RPC
     * will fail with the corresponding [Status].  If it fails with a
     * [java.util.concurrent.CancellationException], the RPC will fail with status
     * `Status.CANCELLED`.  If creating
     * or collecting the returned flow fails for any other reason, the RPC will fail with
     * `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open fun subscribePipelineUpdates(request: ChainerOuterClass.PipelineSubscriptionRequest):
        Flow<ChainerOuterClass.PipelineUpdateMessage> = throw
        StatusException(UNIMPLEMENTED.withDescription("Method seldon.mlops.chainer.Chainer.SubscribePipelineUpdates is unimplemented"))

    /**
     * Returns the response to an RPC for seldon.mlops.chainer.Chainer.PipelineUpdateEvent.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun pipelineUpdateEvent(request: ChainerOuterClass.PipelineUpdateStatusMessage):
        ChainerOuterClass.PipelineUpdateStatusResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method seldon.mlops.chainer.Chainer.PipelineUpdateEvent is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(serverStreamingServerMethodDefinition(
      context = this.context,
      descriptor = ChainerGrpc.getSubscribePipelineUpdatesMethod(),
      implementation = ::subscribePipelineUpdates
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ChainerGrpc.getPipelineUpdateEventMethod(),
      implementation = ::pipelineUpdateEvent
    )).build()
  }
}
