package io.seldon.clustermanager.k8s;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.isNull;
import static org.mockito.Mockito.verify;

import org.junit.Assert;
import org.junit.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Mockito;

import com.google.protobuf.Message;

import io.kubernetes.client.ProgressRequestBody.ProgressRequestListener;
import io.kubernetes.client.ProgressResponseBody.ProgressListener;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;


public class End2EndTest extends End2EndBase {

	@SuppressWarnings("unchecked")
	@Test
	public void testSimpleModel() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments("v1alpha2", 0, 0);

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.times(3)).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.times(2)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.times(1)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));

		// Check individual objects
		ArgumentCaptor<Deployment> argument = ArgumentCaptor.forClass(Deployment.class);
		verify(mockProtoClient,Mockito.times(1)).create(argument.capture(), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		Deployment d = argument.getAllValues().get(0);
		//Assert.assertEquals("test-deployment-fx-market-predictor-svc-orch", d.getMetadata().getName());
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(2, d.getSpec().getTemplate().getSpec().getContainersCount());
		d = argument.getAllValues().get(0);
		Assert.assertTrue(d.getMetadata().getName().startsWith("test-deployment-fx-market-predictor"));
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(2, d.getSpec().getTemplate().getSpec().getContainersCount());
		System.out.println(d.getMetadata().getName());
	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testIgnored() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments("v1alpha2", 1, 1); // version is higher than that in resource so should be ignored

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.never()).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.never()).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.never()).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));
	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testRandomABTest() throws Exception
	{
		createMocks("src/test/resources/random_ab_test.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments("v1alpha2", 0, 0);

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.times(5)).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.times(3)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.times(2)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));
	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testRandomABTest1Pod() throws Exception
	{
		createMocks("src/test/resources/random_ab_test_1pod.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments("v1alpha2", 0, 0);

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.times(4)).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.times(3)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.times(1)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		// Check individual objects
		ArgumentCaptor<Deployment> argument = ArgumentCaptor.forClass(Deployment.class);
		verify(mockProtoClient,Mockito.times(1)).create(argument.capture(), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		Deployment d = argument.getAllValues().get(0);
		System.out.println(d.getMetadata().getName());
		//Assert.assertEquals("test-deployment-abtest-fx-market-predictor-svc-orch", d.getMetadata().getName());
		//Assert.assertEquals(1, d.getSpec().getReplicas());
		//Assert.assertEquals(1, d.getSpec().getTemplate().getSpec().getContainersCount());
		d = argument.getAllValues().get(0);
		Assert.assertTrue(d.getMetadata().getName().startsWith("test-deployment-abtest-fx-market-predictor"));
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(3, d.getSpec().getTemplate().getSpec().getContainersCount());
		System.out.println(d.getMetadata().getName());

	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testInvalidGraph() throws Exception
	{
		createMocks("src/test/resources/model_invalid_graph.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments("v1alpha2", 0, 0); // version is higher than that in resource so should be ignored

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.never()).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.never()).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.never()).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		ArgumentCaptor<byte[]> argument = ArgumentCaptor.forClass(byte[].class);
		verify(mockCustomObjectsApi).replaceNamespacedCustomObject(any(String.class), any(String.class), any(String.class), any(String.class), any(String.class), argument.capture());
		byte[] bytes = argument.getValue();
		SeldonDeployment d = SeldonDeploymentUtils.jsonToSeldonDeployment(new String(bytes));
		Assert.assertEquals("Failed", d.getStatus().getState());
		
	}
	
}
