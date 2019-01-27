package io.seldon.clustermanager.k8s;

import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import org.junit.Assert;
import org.junit.Test;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1ServiceList;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;


public class KubeCRDHandlerImplTest  extends End2EndBase {
	
	@Test
	public void validSeldonDeploymentTest() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockK8sApiProvider, mockK8sClientProvider, props);
		
		SeldonDeployment sDep = crdHandler.getSeldonDeployment("a", "b");
		Assert.assertNotNull(sDep);
	}
	
	@Test
	public void invalidSeldonDeploymentJsonTest() throws Exception
	{
		createMocks("src/test/resources/invalid.json");
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockK8sApiProvider, mockK8sClientProvider, props);
		
		SeldonDeployment sDep = crdHandler.getSeldonDeployment("a", "b");
		Assert.assertNull(sDep);
	}
	
	@Test
	public void invalidSeldonDeploymentTest() throws Exception
	{
		createMocks("src/test/resources/model_invalid.json");
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockK8sApiProvider, mockK8sClientProvider, props);
		
		SeldonDeployment sDep = crdHandler.getSeldonDeployment("a", "b");
		Assert.assertNull(sDep);
	}
	
	
	@Test
	public void getOwnedDeploymentsTest() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		K8sApiProvider mockApiProvider = mock(K8sApiProvider.class);
		ExtensionsV1beta1DeploymentList l = new ExtensionsV1beta1DeploymentList();
		ExtensionsV1beta1Api mockExtensionsApi = mock(ExtensionsV1beta1Api.class);
		when(mockApiProvider.getExtensionsV1beta1Api(any(ApiClient.class))).thenReturn(mockExtensionsApi);
		when(mockExtensionsApi.listNamespacedDeployment(any(String.class), any(String.class), any(String.class), 
				any(String.class), any(Boolean.class), any(String.class), any(Integer.class), 
				any(String.class),any(Integer.class),any(Boolean.class))).thenReturn(l);
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockApiProvider, mockK8sClientProvider, props);
		ExtensionsV1beta1DeploymentList l2 = crdHandler.getOwnedDeployments("a", "b");
		Assert.assertNotNull(l2);
		Assert.assertEquals(l, l2);
	}
	
	@Test
	public void getOwnedDeploymentsExceptionTest() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		K8sApiProvider mockApiProvider = mock(K8sApiProvider.class);
		ExtensionsV1beta1DeploymentList l = new ExtensionsV1beta1DeploymentList();
		ExtensionsV1beta1Api mockExtensionsApi = mock(ExtensionsV1beta1Api.class);
		when(mockApiProvider.getExtensionsV1beta1Api(any(ApiClient.class))).thenReturn(mockExtensionsApi);
		when(mockExtensionsApi.listNamespacedDeployment(any(String.class), any(String.class), any(String.class), 
				any(String.class), any(Boolean.class), any(String.class), any(Integer.class), 
				any(String.class),any(Integer.class),any(Boolean.class))).thenThrow(new ApiException());
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockApiProvider, mockK8sClientProvider, props);
		ExtensionsV1beta1DeploymentList l2 = crdHandler.getOwnedDeployments("a", "b");
		Assert.assertNull(l2);
	}
	
	@Test
	public void getOwnedServicesTest() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		K8sApiProvider mockApiProvider = mock(K8sApiProvider.class);
		V1ServiceList l = new V1ServiceList();
		CoreV1Api mockCoreApi = mock(CoreV1Api.class);
		when(mockApiProvider.getCoreV1Api(any(ApiClient.class))).thenReturn(mockCoreApi);
		when(mockCoreApi.listNamespacedService(any(String.class), any(String.class), any(String.class), 
				any(String.class), any(Boolean.class), any(String.class), any(Integer.class), 
				any(String.class),any(Integer.class),any(Boolean.class))).thenReturn(l);
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockApiProvider, mockK8sClientProvider, props);
		V1ServiceList l2 = crdHandler.getOwnedServices("a", "b");
		Assert.assertNotNull(l2);
		Assert.assertEquals(l, l2);
	}
	
	@Test
	public void getOwnedServicesExceptionTest() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		K8sApiProvider mockApiProvider = mock(K8sApiProvider.class);
		V1ServiceList l = new V1ServiceList();
		CoreV1Api mockCoreApi = mock(CoreV1Api.class);
		when(mockApiProvider.getCoreV1Api(any(ApiClient.class))).thenReturn(mockCoreApi);
		when(mockCoreApi.listNamespacedService(any(String.class), any(String.class), any(String.class), 
				any(String.class), any(Boolean.class), any(String.class), any(Integer.class), 
				any(String.class),any(Integer.class),any(Boolean.class))).thenThrow(new ApiException());
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl(mockApiProvider, mockK8sClientProvider, props);
		V1ServiceList l2 = crdHandler.getOwnedServices("a", "b");
		Assert.assertNull(l2);
	}

}
