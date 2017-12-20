/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
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
