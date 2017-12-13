package io.seldon.apife.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;

public class HeaderServerInterceptor implements ServerInterceptor {

    protected static Logger logger = LoggerFactory.getLogger(HeaderServerInterceptor.class.getName());

    final Metadata.Key<String> CUSTOM_HEADER_KEY =
        Metadata.Key.of("custom_server_header_key", Metadata.ASCII_STRING_MARSHALLER);

    private SeldonGrpcServer server;
    
    public HeaderServerInterceptor(SeldonGrpcServer server) {
      super();
      this.server = server;
  }

  @Override
  public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
        ServerCall<ReqT, RespT> call,
        final Metadata requestHeaders,
        ServerCallHandler<ReqT, RespT> next) {
      logger.info("header received from client:" + requestHeaders);
      return new MessagePrincipalListener<ReqT>(next.startCall(call, requestHeaders),"principal",server);
    }
  }
