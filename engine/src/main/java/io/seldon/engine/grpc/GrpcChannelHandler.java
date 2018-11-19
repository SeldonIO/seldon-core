package io.seldon.engine.grpc;

import java.util.Map;

import org.jboss.netty.util.internal.ConcurrentHashMap;
import org.springframework.stereotype.Component;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.seldon.protos.DeploymentProtos.Endpoint;

@Component
public class GrpcChannelHandler {

	private Map<Endpoint,ManagedChannel> store = new ConcurrentHashMap<>();
	
	public ManagedChannel get(Endpoint endpoint) {
		if (store.containsKey(endpoint))
			return store.get(endpoint);
		else
		{
			ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
			store.putIfAbsent(endpoint, channel);
			return store.get(endpoint);
		}
	}
	
	public int size() {
		return store.size();
	}
	
}
