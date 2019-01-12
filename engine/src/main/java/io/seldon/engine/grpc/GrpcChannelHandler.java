package io.seldon.engine.grpc;

import java.util.Map;

import org.jboss.netty.util.internal.ConcurrentHashMap;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.opentracing.contrib.grpc.ClientTracingInterceptor;
import io.seldon.engine.tracing.TracingProvider;
import io.seldon.protos.DeploymentProtos.Endpoint;

@Component
public class GrpcChannelHandler {

	private Map<Endpoint,Channel> store = new ConcurrentHashMap<>();
	
	@Autowired
	TracingProvider tracingProvider;
	
	public Channel get(Endpoint endpoint) {
		if (store.containsKey(endpoint))
			return store.get(endpoint);
		else
		{
			ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
			
			if (tracingProvider != null && tracingProvider.isActive())
			{
				ClientTracingInterceptor tracingInterceptor = new ClientTracingInterceptor(this.tracingProvider.getTracer());
				store.putIfAbsent(endpoint, tracingInterceptor.intercept(channel));
			}
			else
				store.putIfAbsent(endpoint, channel);
			return store.get(endpoint);
		}
	}
	
	public int size() {
		return store.size();
	}
	
}
