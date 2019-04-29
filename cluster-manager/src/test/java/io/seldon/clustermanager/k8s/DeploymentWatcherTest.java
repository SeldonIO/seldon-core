package io.seldon.clustermanager.k8s;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.isNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Future;

import org.apache.commons.lang3.concurrent.ConcurrentUtils;
import org.junit.Test;
import org.mockito.Mockito;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.squareup.okhttp.Call;
import com.squareup.okhttp.MediaType;
import com.squareup.okhttp.Protocol;
import com.squareup.okhttp.Request;
import com.squareup.okhttp.Response;
import com.squareup.okhttp.ResponseBody;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.JSON;
import io.kubernetes.client.ProgressRequestBody.ProgressRequestListener;
import io.kubernetes.client.ProgressResponseBody.ProgressListener;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.clustermanager.k8s.tasks.K8sTaskScheduler;
import io.seldon.clustermanager.k8s.tasks.SeldonDeploymentTaskKey;

public class DeploymentWatcherTest extends AppTest {

	public static class K8sSynchronousTaskScheduler implements K8sTaskScheduler{

		@Override
		public Future submit(SeldonDeploymentTaskKey key,Runnable task) {
			task.run();
			return ConcurrentUtils.constantFuture(null);
		}
		
	}
	
	SeldonDeploymentStatusUpdate mockStatusUpdater;
	K8sApiProvider mockK8sApiProvider;
	K8sClientProvider mockK8sClientProvider;
	ClusterManagerProperites props;

	@SuppressWarnings("unchecked")
	public void createMocks(String resourceFilename,String watchType) throws Exception
	{
		props = new ClusterManagerProperites();
		props.setEngineContainerImageAndVersion("engine:0.1");
		props.setEngineContainerImagePullPolicy("IfNotPresent");
		props.setEngineContainerServiceAccountName("default");
		
		// Use in watcher
		Request.Builder requestBuilder = new Request.Builder().url("http://0.0.0.0:8000");
		Response.Builder responseBuilder = new Response.Builder();
				String jsonStr = readFile(resourceFilename,StandardCharsets.UTF_8).replaceAll("\n", "") + "\n";
				

		ResponseBody responseBody = ResponseBody.create(MediaType.parse("Content-Type: application/json"), jsonStr);
		Response response = responseBuilder.code(200).request(requestBuilder.build()).protocol(Protocol.HTTP_2).body(responseBody).build();
		Call mockListDeploymentCall = mock(Call.class);
		when(mockListDeploymentCall.execute()).thenReturn(response);
		ExtensionsV1beta1Api mockExtensionApi = mock(ExtensionsV1beta1Api.class);
		when(mockExtensionApi.listNamespacedDeploymentCall(any(String.class), isNull(String.class), isNull(String.class), 
				isNull(String.class),any(Boolean.class),any(String.class), any(Integer.class),
				any(String.class), any(Integer.class), any(Boolean.class),isNull(ProgressListener.class),isNull(ProgressRequestListener.class))).thenReturn(mockListDeploymentCall);
		
		Watch.Response<Object> watchResponse = mock(Watch.Response.class);
		Gson gson = new GsonBuilder().create();
		watchResponse.object = gson.fromJson(jsonStr, ExtensionsV1beta1Deployment.class);
		watchResponse.type = watchType;
		
		// Mock API Client setup - JSON is called from Watch
		JSON mockJSON = mock(JSON.class);
		when(mockJSON.deserialize(any(String.class), any(Type.class))).thenReturn(watchResponse);
		ApiClient mockApiClient = mock(ApiClient.class);
		when(mockApiClient.getJSON()).thenReturn(mockJSON);
		when(mockApiClient.escapeString(any(String.class))).thenCallRealMethod();

		mockK8sClientProvider = mock(K8sClientProvider.class);
		when(mockK8sClientProvider.getClient()).thenReturn(mockApiClient);

		mockK8sApiProvider = mock(K8sApiProvider.class);
		when(mockK8sApiProvider.getExtensionsV1beta1Api(any(ApiClient.class))).thenReturn(mockExtensionApi);
		
		mockStatusUpdater = mock(SeldonDeploymentStatusUpdateImpl.class);
	}
	
	@Test
	public void testAddedDeployment() throws Exception
	{
		createMocks("src/test/resources/deployment_model_deployed.json","ADDED");
		DeploymentWatcher watcher = new DeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, props, mockStatusUpdater, new K8sSynchronousTaskScheduler());
		watcher.watchDeployments(0, 0);
		
		verify(mockStatusUpdater).updateStatus(Mockito.matches("mymodel"), Mockito.matches("v1alpha2"),Mockito.matches("mymodel-mymodel-classifier-0"), Mockito.eq(1), Mockito.eq(1),Mockito.anyString());
	}
	
	@Test
	public void testDeletedDeployment() throws Exception
	{
		createMocks("src/test/resources/deployment_model_deployed.json","DELETED");
		DeploymentWatcher watcher = new DeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, props, mockStatusUpdater, new K8sSynchronousTaskScheduler());
		watcher.watchDeployments(0, 0);
		
		verify(mockStatusUpdater).removeStatus(Mockito.matches("mymodel"), Mockito.matches("v1alpha2"), Mockito.matches("mymodel-mymodel-classifier-0"),Mockito.anyString());
	}
	
	@Test
	public void testSkippedWatch() throws Exception
	{
		createMocks("src/test/resources/deployment_model_deployed.json","DELETED");
		DeploymentWatcher watcher = new DeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, props, mockStatusUpdater, new K8sSynchronousTaskScheduler());
		watcher.watchDeployments(1, 1);
		
		verify(mockStatusUpdater,Mockito.never()).removeStatus(Mockito.matches("mymodel"), Mockito.matches("v1alpha2"), Mockito.matches("mymodel-mymodel-classifier-0"),Mockito.anyString());
	}
}
