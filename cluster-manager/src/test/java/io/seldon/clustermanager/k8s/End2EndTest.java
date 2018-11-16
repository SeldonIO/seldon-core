package io.seldon.clustermanager.k8s;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.isNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Mockito;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.internal.LinkedTreeMap;
import com.google.protobuf.Any.Builder;
import com.google.protobuf.Message;
import com.squareup.okhttp.Call;
import com.squareup.okhttp.MediaType;
import com.squareup.okhttp.Protocol;
import com.squareup.okhttp.Request;
import com.squareup.okhttp.Response;
import com.squareup.okhttp.ResponseBody;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.JSON;
import io.kubernetes.client.ProgressRequestBody.ProgressRequestListener;
import io.kubernetes.client.ProgressResponseBody.ProgressListener;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.proto.Meta.Status;
import io.kubernetes.client.proto.V1;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;



//@RunWith(PowerMockRunner.class)
//@PrepareForTest({CustomObjectsApi.class,KubeCRDHandlerImpl.class})
public class End2EndTest extends AppTest {

	K8sClientProvider mockK8sClientProvider;
	ProtoClient mockProtoClient;
	CRDCreator mockCrdCreator;
	CustomObjectsApi mockCustomObjectsApi;
	K8sApiProvider mockK8sApiProvider;
	
	KubeCRDHandler crdHandler;
	ClusterManagerProperites props;
	SeldonDeploymentCache mlCache;
	SeldonDeploymentOperatorImpl operator;
	SeldonDeploymentControllerImpl controller;
	
	
	@SuppressWarnings("unchecked")
	public void createMocks(String resourceFilename) throws Exception
	{
		ApiClient client = new ApiClient();
		
		// Allow use to mock the API client that is created
		//PowerMockito.mockStatic(Config.class);
		//Mockito.when(Config.defaultClient()).thenReturn(client);
		
		// Handle the CRD creator and make it show the CRD exists
		ApiException k8sException = new ApiException(409, "hahaha");
		mockCrdCreator = mock(CRDCreator.class);
		Mockito.doNothing().when(mockCrdCreator).createCRD();
		//when(mockCrdCreator.createCustomResourceDefinition(any(ApiClient.class),any(byte[].class), isNull(String.class))).thenThrow(k8sException);
		//doCallRealMethod().when(mockCrdCreator).createCRD();
		
		// Handle watches
		/*
		Watch<Object> watch = mock(Watch.class);
		PowerMockito.mockStatic(Watch.class);
		Mockito.when(Watch.createWatch(any(ApiClient.class), any(Call.class), any(Type.class))).thenReturn(watch); //return our mock watch when asked
		// Mock a watch response
		Watch.Response<Object> watchResponse = mock(Watch.Response.class);
		String jsonStr = readFile(resourceFilename,StandardCharsets.UTF_8);
		Gson gson = new GsonBuilder().create();
		watchResponse.object = gson.fromJson(jsonStr, LinkedTreeMap.class);
		watchResponse.type = "ADDED";
		// ensure its returned once
		Mockito.when(watch.hasNext()).thenReturn(true,false);
		Mockito.when(watch.next()).thenReturn(watchResponse);
		Mockito.when(watch.iterator()).thenReturn(watch);
		*/
		
		// Handle the call to CustomObjectsApi
		mockCustomObjectsApi = mock(CustomObjectsApi.class);
		//whenNew(CustomObjectsApi.class).withAnyArguments().thenReturn(mockCustomObjectsApi);
		// Use in watcher
		Request.Builder requestBuilder = new Request.Builder().url("http://0.0.0.0:8000");
		Response.Builder responseBuilder = new Response.Builder();
		String jsonStr = readFile(resourceFilename,StandardCharsets.UTF_8).replaceAll("\n", "") + "\n";
		

		ResponseBody responseBody = ResponseBody.create(MediaType.parse("Content-Type: application/json"), jsonStr);
		Response response = responseBuilder.code(200).request(requestBuilder.build()).protocol(Protocol.HTTP_2).body(responseBody).build();
		Call mockListNamespaceCall = mock(Call.class);
		when(mockListNamespaceCall.execute()).thenReturn(response);
		// use in watcher
		when(mockCustomObjectsApi.listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class))).thenReturn(mockListNamespaceCall);
		// Use in crdHandler
		when(mockCustomObjectsApi.replaceNamespacedCustomObjectStatus(any(String.class), any(String.class), any(String.class), any(String.class), any(String.class), any(byte[].class))).thenThrow(ApiException.class);
		when(mockCustomObjectsApi.replaceNamespacedCustomObject(any(String.class), any(String.class), any(String.class), any(String.class), any(String.class), any(byte[].class))).thenReturn(null);		
		
		mockK8sApiProvider = mock(K8sApiProvider.class);
		when(mockK8sApiProvider.getCustomObjectsApi(any(ApiClient.class))).thenReturn(mockCustomObjectsApi);
		
		Watch.Response<Object> watchResponse = mock(Watch.Response.class);
		Gson gson = new GsonBuilder().create();
		watchResponse.object = gson.fromJson(jsonStr, LinkedTreeMap.class);
		watchResponse.type = "ADDED";
		
		JSON mockJSON = mock(JSON.class);
		when(mockJSON.deserialize(any(String.class), any(Type.class))).thenReturn(watchResponse);
		ApiClient mockApiClient = mock(ApiClient.class);
		when(mockApiClient.getJSON()).thenReturn(mockJSON);
		when(mockApiClient.escapeString(any(String.class))).thenCallRealMethod();

		
		mockProtoClient = mock(ProtoClient.class);
		Status status404 = Status.newBuilder().setCode(404).build();
		V1.Container dummyResponseObject = V1.Container.newBuilder().setName("DummyObject").build();
		ObjectOrStatus objectOrstatus404 = new ObjectOrStatus(null,status404);
		ObjectOrStatus objectOrstatusOk = new ObjectOrStatus(dummyResponseObject,null);
		when(mockProtoClient.update(any(Message.class), any(String.class), any(String.class), any(String.class))).thenReturn(objectOrstatusOk);
		when(mockProtoClient.create(any(Message.class), any(String.class), any(String.class), any(String.class))).thenReturn(objectOrstatusOk);
		when(mockProtoClient.list(any(Builder.class), any(String.class))).thenReturn(objectOrstatus404);
		when(mockProtoClient.getApiClient()).thenReturn(client);

		
		mockK8sClientProvider = mock(K8sClientProvider.class);
		when(mockK8sClientProvider.getClient()).thenReturn(mockApiClient);
		when(mockK8sClientProvider.getProtoClient()).thenReturn(mockProtoClient);

		props = new ClusterManagerProperites();
		props.setEngineContainerImageAndVersion("engine:0.1");
		props.setEngineContainerImagePullPolicy("IfNotPresent");
		props.setEngineContainerServiceAccountName("default");
		crdHandler = new KubeCRDHandlerImpl(mockK8sApiProvider,mockK8sClientProvider,props);
		mlCache = new SeldonDeploymentCacheImpl(props, crdHandler);
		operator = new SeldonDeploymentOperatorImpl(props);
		controller = new SeldonDeploymentControllerImpl(operator, mockK8sClientProvider, crdHandler, mlCache);
		
	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testSimpleModel() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments(0, 0);

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.times(4)).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.times(2)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.times(2)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));

		// Check individual objects
		ArgumentCaptor<Deployment> argument = ArgumentCaptor.forClass(Deployment.class);
		verify(mockProtoClient,Mockito.times(2)).create(argument.capture(), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		Deployment d = argument.getAllValues().get(0);
		Assert.assertEquals("test-deployment-fx-market-predictor-svc-orch", d.getMetadata().getName());
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(1, d.getSpec().getTemplate().getSpec().getContainersCount());
		d = argument.getAllValues().get(1);
		Assert.assertTrue(d.getMetadata().getName().startsWith("test-deployment-fx-market-predictor"));
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(1, d.getSpec().getTemplate().getSpec().getContainersCount());
		System.out.println(d.getMetadata().getName());
	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testIgnored() throws Exception
	{
		createMocks("src/test/resources/model_simple.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments(1, 1); // version is higher than that in resource so should be ignored

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
		sdWatcher.watchSeldonMLDeployments(0, 0);

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.times(6)).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.times(3)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.times(3)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));
	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testRandomABTest1Pod() throws Exception
	{
		createMocks("src/test/resources/random_ab_test_1pod.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments(0, 0);

		verify(mockCustomObjectsApi).listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class), 
				any(String.class), isNull(String.class), isNull(String.class), any(String.class), 
				any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class));
		verify(mockProtoClient,Mockito.times(5)).create(any(Message.class), any(String.class), any(String.class), any(String.class));
		verify(mockProtoClient,Mockito.times(3)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Service"));
		verify(mockProtoClient,Mockito.times(2)).create(any(Message.class), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		// Check individual objects
		ArgumentCaptor<Deployment> argument = ArgumentCaptor.forClass(Deployment.class);
		verify(mockProtoClient,Mockito.times(2)).create(argument.capture(), any(String.class), any(String.class), Mockito.matches("Deployment"));
		
		Deployment d = argument.getAllValues().get(0);
		System.out.println(d.getMetadata().getName());
		Assert.assertEquals("test-deployment-abtest-fx-market-predictor-svc-orch", d.getMetadata().getName());
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(1, d.getSpec().getTemplate().getSpec().getContainersCount());
		d = argument.getAllValues().get(1);
		Assert.assertTrue(d.getMetadata().getName().startsWith("test-deployment-abtest-fx-market-predictor"));
		Assert.assertEquals(1, d.getSpec().getReplicas());
		Assert.assertEquals(2, d.getSpec().getTemplate().getSpec().getContainersCount());
		System.out.println(d.getMetadata().getName());

	}
	
	@SuppressWarnings("unchecked")
	@Test
	public void testInvalidGraph() throws Exception
	{
		createMocks("src/test/resources/model_invalid_graph.json");
		SeldonDeploymentWatcher sdWatcher = new SeldonDeploymentWatcher(mockK8sApiProvider,mockK8sClientProvider,mockCrdCreator, props, controller, mlCache, crdHandler);
		sdWatcher.watchSeldonMLDeployments(0, 0); // version is higher than that in resource so should be ignored

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
