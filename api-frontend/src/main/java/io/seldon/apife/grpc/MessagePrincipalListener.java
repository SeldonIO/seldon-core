package io.seldon.apife.grpc;

import io.grpc.ForwardingServerCallListener;
import io.grpc.ServerCall;
import io.grpc.ServerCall.Listener;

public class MessagePrincipalListener <R> extends ForwardingServerCallListener<R>
{
    ServerCall.Listener<R> delegate;
    SeldonGrpcServer server;
    String principal;
    
    public MessagePrincipalListener(ServerCall.Listener<R> delegate,String principal,SeldonGrpcServer server) {
        this.delegate = delegate;
        this.server = server;
        this.principal = principal;
    }
    
    @Override
    protected Listener<R> delegate() {
        return delegate;
    }
    
    @Override
    public void onMessage(R request) {
        server.setPrincipal(this.principal);
        super.onMessage(request);
    }
    
}
