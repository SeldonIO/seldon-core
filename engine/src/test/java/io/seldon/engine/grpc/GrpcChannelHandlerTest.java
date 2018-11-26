package io.seldon.engine.grpc;

import org.junit.Assert;
import org.junit.Test;

import io.grpc.ManagedChannel;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.Endpoint.EndpointType;




public class GrpcChannelHandlerTest {
	
	@Test
	public void testIdentity()
	{
		GrpcChannelHandler ch = new GrpcChannelHandler();
		
		Endpoint e1 = Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).build();
		Endpoint e2 = Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).build();
		
		ManagedChannel mc1 = ch.get(e1);
		ManagedChannel mc2 = ch.get(e2);
		
		Assert.assertEquals(mc1, mc2);
		Assert.assertEquals(1, ch.size());
	}

	@Test
	public void testDifferenceByType()
	{
		GrpcChannelHandler ch = new GrpcChannelHandler();
		
		Endpoint e1 = Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).setType(EndpointType.REST).build();
		Endpoint e2 = Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).setType(EndpointType.GRPC).build();
		
		ManagedChannel mc1 = ch.get(e1);
		ManagedChannel mc2 = ch.get(e2);
		
		Assert.assertNotEquals(mc1, mc2);
		Assert.assertEquals(2, ch.size());
	}
	
	@Test
	public void testDifferenceByHost()
	{
		GrpcChannelHandler ch = new GrpcChannelHandler();
		
		Endpoint e1 = Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).setType(EndpointType.REST).build();
		Endpoint e2 = Endpoint.newBuilder().setServiceHost("hostB").setServicePort(1000).setType(EndpointType.REST).build();
		
		ManagedChannel mc1 = ch.get(e1);
		ManagedChannel mc2 = ch.get(e2);
		
		Assert.assertNotEquals(mc1, mc2);
		Assert.assertEquals(2, ch.size());
	}

}
