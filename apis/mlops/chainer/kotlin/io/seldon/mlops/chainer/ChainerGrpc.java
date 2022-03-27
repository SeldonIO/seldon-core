package io.seldon.mlops.chainer;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.44.1)",
    comments = "Source: chainer.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ChainerGrpc {

  private ChainerGrpc() {}

  public static final String SERVICE_NAME = "seldon.mlops.chainer.Chainer";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest,
      io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage> getSubscribePipelineUpdatesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SubscribePipelineUpdates",
      requestType = io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest.class,
      responseType = io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest,
      io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage> getSubscribePipelineUpdatesMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest, io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage> getSubscribePipelineUpdatesMethod;
    if ((getSubscribePipelineUpdatesMethod = ChainerGrpc.getSubscribePipelineUpdatesMethod) == null) {
      synchronized (ChainerGrpc.class) {
        if ((getSubscribePipelineUpdatesMethod = ChainerGrpc.getSubscribePipelineUpdatesMethod) == null) {
          ChainerGrpc.getSubscribePipelineUpdatesMethod = getSubscribePipelineUpdatesMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest, io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SubscribePipelineUpdates"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.getDefaultInstance()))
              .setSchemaDescriptor(new ChainerMethodDescriptorSupplier("SubscribePipelineUpdates"))
              .build();
        }
      }
    }
    return getSubscribePipelineUpdatesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage,
      io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse> getPipelineUpdateEventMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "PipelineUpdateEvent",
      requestType = io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage.class,
      responseType = io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage,
      io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse> getPipelineUpdateEventMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage, io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse> getPipelineUpdateEventMethod;
    if ((getPipelineUpdateEventMethod = ChainerGrpc.getPipelineUpdateEventMethod) == null) {
      synchronized (ChainerGrpc.class) {
        if ((getPipelineUpdateEventMethod = ChainerGrpc.getPipelineUpdateEventMethod) == null) {
          ChainerGrpc.getPipelineUpdateEventMethod = getPipelineUpdateEventMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage, io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "PipelineUpdateEvent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ChainerMethodDescriptorSupplier("PipelineUpdateEvent"))
              .build();
        }
      }
    }
    return getPipelineUpdateEventMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ChainerStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ChainerStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ChainerStub>() {
        @java.lang.Override
        public ChainerStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ChainerStub(channel, callOptions);
        }
      };
    return ChainerStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ChainerBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ChainerBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ChainerBlockingStub>() {
        @java.lang.Override
        public ChainerBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ChainerBlockingStub(channel, callOptions);
        }
      };
    return ChainerBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ChainerFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ChainerFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ChainerFutureStub>() {
        @java.lang.Override
        public ChainerFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ChainerFutureStub(channel, callOptions);
        }
      };
    return ChainerFutureStub.newStub(factory, channel);
  }

  /**
   */
  public static abstract class ChainerImplBase implements io.grpc.BindableService {

    /**
     */
    public void subscribePipelineUpdates(io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSubscribePipelineUpdatesMethod(), responseObserver);
    }

    /**
     */
    public void pipelineUpdateEvent(io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPipelineUpdateEventMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getSubscribePipelineUpdatesMethod(),
            io.grpc.stub.ServerCalls.asyncServerStreamingCall(
              new MethodHandlers<
                io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest,
                io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage>(
                  this, METHODID_SUBSCRIBE_PIPELINE_UPDATES)))
          .addMethod(
            getPipelineUpdateEventMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage,
                io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse>(
                  this, METHODID_PIPELINE_UPDATE_EVENT)))
          .build();
    }
  }

  /**
   */
  public static final class ChainerStub extends io.grpc.stub.AbstractAsyncStub<ChainerStub> {
    private ChainerStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ChainerStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ChainerStub(channel, callOptions);
    }

    /**
     */
    public void subscribePipelineUpdates(io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getSubscribePipelineUpdatesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void pipelineUpdateEvent(io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPipelineUpdateEventMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   */
  public static final class ChainerBlockingStub extends io.grpc.stub.AbstractBlockingStub<ChainerBlockingStub> {
    private ChainerBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ChainerBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ChainerBlockingStub(channel, callOptions);
    }

    /**
     */
    public java.util.Iterator<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage> subscribePipelineUpdates(
        io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getSubscribePipelineUpdatesMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse pipelineUpdateEvent(io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPipelineUpdateEventMethod(), getCallOptions(), request);
    }
  }

  /**
   */
  public static final class ChainerFutureStub extends io.grpc.stub.AbstractFutureStub<ChainerFutureStub> {
    private ChainerFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ChainerFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ChainerFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse> pipelineUpdateEvent(
        io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPipelineUpdateEventMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_SUBSCRIBE_PIPELINE_UPDATES = 0;
  private static final int METHODID_PIPELINE_UPDATE_EVENT = 1;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final ChainerImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(ChainerImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_SUBSCRIBE_PIPELINE_UPDATES:
          serviceImpl.subscribePipelineUpdates((io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage>) responseObserver);
          break;
        case METHODID_PIPELINE_UPDATE_EVENT:
          serviceImpl.pipelineUpdateEvent((io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  private static abstract class ChainerBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ChainerBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.seldon.mlops.chainer.ChainerOuterClass.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Chainer");
    }
  }

  private static final class ChainerFileDescriptorSupplier
      extends ChainerBaseDescriptorSupplier {
    ChainerFileDescriptorSupplier() {}
  }

  private static final class ChainerMethodDescriptorSupplier
      extends ChainerBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    ChainerMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (ChainerGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ChainerFileDescriptorSupplier())
              .addMethod(getSubscribePipelineUpdatesMethod())
              .addMethod(getPipelineUpdateEventMethod())
              .build();
        }
      }
    }
    return result;
  }
}
