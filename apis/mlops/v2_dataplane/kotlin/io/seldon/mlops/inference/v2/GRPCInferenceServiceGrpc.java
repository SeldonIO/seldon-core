package io.seldon.mlops.inference.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Inference Server GRPC endpoints.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.44.1)",
    comments = "Source: v2_dataplane.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class GRPCInferenceServiceGrpc {

  private GRPCInferenceServiceGrpc() {}

  public static final String SERVICE_NAME = "inference.GRPCInferenceService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse> getServerLiveMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ServerLive",
      requestType = io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest.class,
      responseType = io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse> getServerLiveMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest, io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse> getServerLiveMethod;
    if ((getServerLiveMethod = GRPCInferenceServiceGrpc.getServerLiveMethod) == null) {
      synchronized (GRPCInferenceServiceGrpc.class) {
        if ((getServerLiveMethod = GRPCInferenceServiceGrpc.getServerLiveMethod) == null) {
          GRPCInferenceServiceGrpc.getServerLiveMethod = getServerLiveMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest, io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ServerLive"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse.getDefaultInstance()))
              .setSchemaDescriptor(new GRPCInferenceServiceMethodDescriptorSupplier("ServerLive"))
              .build();
        }
      }
    }
    return getServerLiveMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse> getServerReadyMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ServerReady",
      requestType = io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest.class,
      responseType = io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse> getServerReadyMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest, io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse> getServerReadyMethod;
    if ((getServerReadyMethod = GRPCInferenceServiceGrpc.getServerReadyMethod) == null) {
      synchronized (GRPCInferenceServiceGrpc.class) {
        if ((getServerReadyMethod = GRPCInferenceServiceGrpc.getServerReadyMethod) == null) {
          GRPCInferenceServiceGrpc.getServerReadyMethod = getServerReadyMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest, io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ServerReady"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse.getDefaultInstance()))
              .setSchemaDescriptor(new GRPCInferenceServiceMethodDescriptorSupplier("ServerReady"))
              .build();
        }
      }
    }
    return getServerReadyMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse> getModelReadyMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ModelReady",
      requestType = io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest.class,
      responseType = io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse> getModelReadyMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest, io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse> getModelReadyMethod;
    if ((getModelReadyMethod = GRPCInferenceServiceGrpc.getModelReadyMethod) == null) {
      synchronized (GRPCInferenceServiceGrpc.class) {
        if ((getModelReadyMethod = GRPCInferenceServiceGrpc.getModelReadyMethod) == null) {
          GRPCInferenceServiceGrpc.getModelReadyMethod = getModelReadyMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest, io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ModelReady"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse.getDefaultInstance()))
              .setSchemaDescriptor(new GRPCInferenceServiceMethodDescriptorSupplier("ModelReady"))
              .build();
        }
      }
    }
    return getModelReadyMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse> getServerMetadataMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ServerMetadata",
      requestType = io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest.class,
      responseType = io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse> getServerMetadataMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest, io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse> getServerMetadataMethod;
    if ((getServerMetadataMethod = GRPCInferenceServiceGrpc.getServerMetadataMethod) == null) {
      synchronized (GRPCInferenceServiceGrpc.class) {
        if ((getServerMetadataMethod = GRPCInferenceServiceGrpc.getServerMetadataMethod) == null) {
          GRPCInferenceServiceGrpc.getServerMetadataMethod = getServerMetadataMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest, io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ServerMetadata"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse.getDefaultInstance()))
              .setSchemaDescriptor(new GRPCInferenceServiceMethodDescriptorSupplier("ServerMetadata"))
              .build();
        }
      }
    }
    return getServerMetadataMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse> getModelMetadataMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ModelMetadata",
      requestType = io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest.class,
      responseType = io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse> getModelMetadataMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest, io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse> getModelMetadataMethod;
    if ((getModelMetadataMethod = GRPCInferenceServiceGrpc.getModelMetadataMethod) == null) {
      synchronized (GRPCInferenceServiceGrpc.class) {
        if ((getModelMetadataMethod = GRPCInferenceServiceGrpc.getModelMetadataMethod) == null) {
          GRPCInferenceServiceGrpc.getModelMetadataMethod = getModelMetadataMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest, io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ModelMetadata"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse.getDefaultInstance()))
              .setSchemaDescriptor(new GRPCInferenceServiceMethodDescriptorSupplier("ModelMetadata"))
              .build();
        }
      }
    }
    return getModelMetadataMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse> getModelInferMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ModelInfer",
      requestType = io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest.class,
      responseType = io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest,
      io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse> getModelInferMethod() {
    io.grpc.MethodDescriptor<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest, io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse> getModelInferMethod;
    if ((getModelInferMethod = GRPCInferenceServiceGrpc.getModelInferMethod) == null) {
      synchronized (GRPCInferenceServiceGrpc.class) {
        if ((getModelInferMethod = GRPCInferenceServiceGrpc.getModelInferMethod) == null) {
          GRPCInferenceServiceGrpc.getModelInferMethod = getModelInferMethod =
              io.grpc.MethodDescriptor.<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest, io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ModelInfer"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse.getDefaultInstance()))
              .setSchemaDescriptor(new GRPCInferenceServiceMethodDescriptorSupplier("ModelInfer"))
              .build();
        }
      }
    }
    return getModelInferMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static GRPCInferenceServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<GRPCInferenceServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<GRPCInferenceServiceStub>() {
        @java.lang.Override
        public GRPCInferenceServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new GRPCInferenceServiceStub(channel, callOptions);
        }
      };
    return GRPCInferenceServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static GRPCInferenceServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<GRPCInferenceServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<GRPCInferenceServiceBlockingStub>() {
        @java.lang.Override
        public GRPCInferenceServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new GRPCInferenceServiceBlockingStub(channel, callOptions);
        }
      };
    return GRPCInferenceServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static GRPCInferenceServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<GRPCInferenceServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<GRPCInferenceServiceFutureStub>() {
        @java.lang.Override
        public GRPCInferenceServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new GRPCInferenceServiceFutureStub(channel, callOptions);
        }
      };
    return GRPCInferenceServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Inference Server GRPC endpoints.
   * </pre>
   */
  public static abstract class GRPCInferenceServiceImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * The ServerLive API indicates if the inference server is able to receive 
     * and respond to metadata and inference requests.
     * </pre>
     */
    public void serverLive(io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getServerLiveMethod(), responseObserver);
    }

    /**
     * <pre>
     * The ServerReady API indicates if the server is ready for inferencing.
     * </pre>
     */
    public void serverReady(io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getServerReadyMethod(), responseObserver);
    }

    /**
     * <pre>
     * The ModelReady API indicates if a specific model is ready for inferencing.
     * </pre>
     */
    public void modelReady(io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getModelReadyMethod(), responseObserver);
    }

    /**
     * <pre>
     * The ServerMetadata API provides information about the server. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public void serverMetadata(io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getServerMetadataMethod(), responseObserver);
    }

    /**
     * <pre>
     * The per-model metadata API provides information about a model. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public void modelMetadata(io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getModelMetadataMethod(), responseObserver);
    }

    /**
     * <pre>
     * The ModelInfer API performs inference using the specified model. Errors are
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public void modelInfer(io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getModelInferMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getServerLiveMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest,
                io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse>(
                  this, METHODID_SERVER_LIVE)))
          .addMethod(
            getServerReadyMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest,
                io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse>(
                  this, METHODID_SERVER_READY)))
          .addMethod(
            getModelReadyMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest,
                io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse>(
                  this, METHODID_MODEL_READY)))
          .addMethod(
            getServerMetadataMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest,
                io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse>(
                  this, METHODID_SERVER_METADATA)))
          .addMethod(
            getModelMetadataMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest,
                io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse>(
                  this, METHODID_MODEL_METADATA)))
          .addMethod(
            getModelInferMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest,
                io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse>(
                  this, METHODID_MODEL_INFER)))
          .build();
    }
  }

  /**
   * <pre>
   * Inference Server GRPC endpoints.
   * </pre>
   */
  public static final class GRPCInferenceServiceStub extends io.grpc.stub.AbstractAsyncStub<GRPCInferenceServiceStub> {
    private GRPCInferenceServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected GRPCInferenceServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new GRPCInferenceServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * The ServerLive API indicates if the inference server is able to receive 
     * and respond to metadata and inference requests.
     * </pre>
     */
    public void serverLive(io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getServerLiveMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * The ServerReady API indicates if the server is ready for inferencing.
     * </pre>
     */
    public void serverReady(io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getServerReadyMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * The ModelReady API indicates if a specific model is ready for inferencing.
     * </pre>
     */
    public void modelReady(io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getModelReadyMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * The ServerMetadata API provides information about the server. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public void serverMetadata(io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getServerMetadataMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * The per-model metadata API provides information about a model. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public void modelMetadata(io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getModelMetadataMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * The ModelInfer API performs inference using the specified model. Errors are
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public void modelInfer(io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest request,
        io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getModelInferMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Inference Server GRPC endpoints.
   * </pre>
   */
  public static final class GRPCInferenceServiceBlockingStub extends io.grpc.stub.AbstractBlockingStub<GRPCInferenceServiceBlockingStub> {
    private GRPCInferenceServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected GRPCInferenceServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new GRPCInferenceServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * The ServerLive API indicates if the inference server is able to receive 
     * and respond to metadata and inference requests.
     * </pre>
     */
    public io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse serverLive(io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getServerLiveMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * The ServerReady API indicates if the server is ready for inferencing.
     * </pre>
     */
    public io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse serverReady(io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getServerReadyMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * The ModelReady API indicates if a specific model is ready for inferencing.
     * </pre>
     */
    public io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse modelReady(io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getModelReadyMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * The ServerMetadata API provides information about the server. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse serverMetadata(io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getServerMetadataMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * The per-model metadata API provides information about a model. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse modelMetadata(io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getModelMetadataMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * The ModelInfer API performs inference using the specified model. Errors are
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse modelInfer(io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getModelInferMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Inference Server GRPC endpoints.
   * </pre>
   */
  public static final class GRPCInferenceServiceFutureStub extends io.grpc.stub.AbstractFutureStub<GRPCInferenceServiceFutureStub> {
    private GRPCInferenceServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected GRPCInferenceServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new GRPCInferenceServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * The ServerLive API indicates if the inference server is able to receive 
     * and respond to metadata and inference requests.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse> serverLive(
        io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getServerLiveMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * The ServerReady API indicates if the server is ready for inferencing.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse> serverReady(
        io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getServerReadyMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * The ModelReady API indicates if a specific model is ready for inferencing.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse> modelReady(
        io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getModelReadyMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * The ServerMetadata API provides information about the server. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse> serverMetadata(
        io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getServerMetadataMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * The per-model metadata API provides information about a model. Errors are 
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse> modelMetadata(
        io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getModelMetadataMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * The ModelInfer API performs inference using the specified model. Errors are
     * indicated by the google.rpc.Status returned for the request. The OK code 
     * indicates success and other codes indicate failure.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse> modelInfer(
        io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getModelInferMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_SERVER_LIVE = 0;
  private static final int METHODID_SERVER_READY = 1;
  private static final int METHODID_MODEL_READY = 2;
  private static final int METHODID_SERVER_METADATA = 3;
  private static final int METHODID_MODEL_METADATA = 4;
  private static final int METHODID_MODEL_INFER = 5;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final GRPCInferenceServiceImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(GRPCInferenceServiceImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_SERVER_LIVE:
          serviceImpl.serverLive((io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerLiveResponse>) responseObserver);
          break;
        case METHODID_SERVER_READY:
          serviceImpl.serverReady((io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerReadyResponse>) responseObserver);
          break;
        case METHODID_MODEL_READY:
          serviceImpl.modelReady((io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelReadyResponse>) responseObserver);
          break;
        case METHODID_SERVER_METADATA:
          serviceImpl.serverMetadata((io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse>) responseObserver);
          break;
        case METHODID_MODEL_METADATA:
          serviceImpl.modelMetadata((io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelMetadataResponse>) responseObserver);
          break;
        case METHODID_MODEL_INFER:
          serviceImpl.modelInfer((io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest) request,
              (io.grpc.stub.StreamObserver<io.seldon.mlops.inference.v2.V2Dataplane.ModelInferResponse>) responseObserver);
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

  private static abstract class GRPCInferenceServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    GRPCInferenceServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.seldon.mlops.inference.v2.V2Dataplane.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("GRPCInferenceService");
    }
  }

  private static final class GRPCInferenceServiceFileDescriptorSupplier
      extends GRPCInferenceServiceBaseDescriptorSupplier {
    GRPCInferenceServiceFileDescriptorSupplier() {}
  }

  private static final class GRPCInferenceServiceMethodDescriptorSupplier
      extends GRPCInferenceServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    GRPCInferenceServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (GRPCInferenceServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new GRPCInferenceServiceFileDescriptorSupplier())
              .addMethod(getServerLiveMethod())
              .addMethod(getServerReadyMethod())
              .addMethod(getModelReadyMethod())
              .addMethod(getServerMetadataMethod())
              .addMethod(getModelMetadataMethod())
              .addMethod(getModelInferMethod())
              .build();
        }
      }
    }
    return result;
  }
}
