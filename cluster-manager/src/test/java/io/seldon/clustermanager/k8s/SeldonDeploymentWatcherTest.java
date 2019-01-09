package io.seldon.clustermanager.k8s;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.squareup.okhttp.*;
import io.kubernetes.client.ApiClient;
import io.kubernetes.client.JSON;
import io.kubernetes.client.ProgressRequestBody;
import io.kubernetes.client.ProgressResponseBody;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.protos.DeploymentProtos;
import org.junit.Assert;
import org.junit.Test;

import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.isNull;
import static org.mockito.Mockito.*;

public class SeldonDeploymentWatcherTest extends AppTest {

    private SeldonDeploymentController mockSeldonDeploymentController;
    private SeldonDeploymentCache mockMLCache;
    private ClusterManagerProperites props;
    private KubeCRDHandler mockCRDHandler;
    private K8sClientProvider mockK8sClientProvider;
    private K8sApiProvider mockK8sApiProvider;
    private CRDCreator mockCRDCreator;

    public void createMocks(String resourceFilename, String watchType,boolean isSingleNameSpace) throws Exception {
        props = new ClusterManagerProperites();
        props.setEngineContainerImageAndVersion("engine:0.1");
        props.setEngineContainerImagePullPolicy("IfNotPresent");
        props.setEngineContainerServiceAccountName("default");
        props.setSingleNamespace(isSingleNameSpace);


        // Use in watcher
        Request.Builder requestBuilder = new Request.Builder().url("http://0.0.0.0:8000");
        Response.Builder responseBuilder = new Response.Builder();
        String jsonStr = readFile(resourceFilename, StandardCharsets.UTF_8).replaceAll("\n", "") + "\n";

        ResponseBody responseBody = ResponseBody.create(MediaType.parse("Content-Type: application/json"), jsonStr);
        Response response = responseBuilder.code(200).request(requestBuilder.build()).protocol(Protocol.HTTP_2).body(responseBody).build();
        Call mockCall = mock(Call.class);
        when(mockCall.execute()).thenReturn(response);
        ExtensionsV1beta1Api mockExtensionApi = mock(ExtensionsV1beta1Api.class);

        when(mockExtensionApi.listNamespacedDeploymentCall(any(String.class), isNull(String.class), isNull(String.class),
                isNull(String.class), any(Boolean.class), any(String.class), any(Integer.class),
                any(String.class), any(Integer.class), any(Boolean.class), isNull(ProgressResponseBody.ProgressListener.class), isNull(ProgressRequestBody.ProgressRequestListener.class))).thenReturn(mockCall);
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

        CustomObjectsApi mockCustomObjectApi = mock(CustomObjectsApi.class);

        when(mockCustomObjectApi.listNamespacedCustomObjectCall(any(String.class), any(String.class), any(String.class),
                any(String.class), any(String.class), any(String.class), any(String.class),
                any(Boolean.class), isNull(ProgressResponseBody.ProgressListener.class), isNull(ProgressRequestBody.ProgressRequestListener.class)
        )).thenReturn(mockCall);

        when(mockCustomObjectApi.listClusterCustomObjectCall(any(String.class), any(String.class), any(String.class),
                any(String.class), any(String.class), any(String.class), any(Boolean.class), isNull(ProgressResponseBody.ProgressListener.class),
                isNull(ProgressRequestBody.ProgressRequestListener.class)
        )).thenReturn(mockCall);

        mockK8sClientProvider = mock(K8sClientProvider.class);
        when(mockK8sClientProvider.getClient()).thenReturn(mockApiClient);

        mockK8sApiProvider = mock(K8sApiProvider.class);
        when(mockK8sApiProvider.getExtensionsV1beta1Api(any(ApiClient.class))).thenReturn(mockExtensionApi);

        when(mockK8sApiProvider.getCustomObjectsApi(mockApiClient)).thenReturn(mockCustomObjectApi);

        mockSeldonDeploymentController = mock(SeldonDeploymentController.class);

        mockCRDCreator = mock(CRDCreator.class);
        doNothing().when(mockCRDCreator).createCRD();

        mockCRDHandler = mock(KubeCRDHandler.class);

        mockMLCache = mock(SeldonDeploymentCache.class);
    }

    @Test
    public void testAddedSeldonMLDeploymentsForSingleNamespaceCluster_SuccessCase() throws Exception {

        createMocks("src/test/resources/model_simple.json", "ADDED",true);
        props.setSingleNamespace(true);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);

        Assert.assertEquals("Expected and Actual resource versions do not match !!!",
                1, watcher.watchSeldonMLDeployments(0, 0));
        verify(mockSeldonDeploymentController,times(1)).createOrReplaceSeldonDeployment(any(DeploymentProtos.SeldonDeployment.class));
    }

    @Test
    public void testModifiedSeldonMLDeploymentsForSingleNamespaceCluster() throws Exception {

        createMocks("src/test/resources/model_simple.json", "MODIFIED",true);
        props.setSingleNamespace(true);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        watcher.watchSeldonMLDeployments(0, 0);
        verify(mockSeldonDeploymentController,times(1)).createOrReplaceSeldonDeployment(any(DeploymentProtos.SeldonDeployment.class));
    }

    @Test
    public void testDeletedSeldonMLDeploymentsForSingleNameSpaceCluster() throws Exception {
        createMocks("src/test/resources/model_simple.json", "DELETED",true);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        watcher.watchSeldonMLDeployments(0,0);
        verify(mockMLCache,times(1)).remove(any(DeploymentProtos.SeldonDeployment.class));
    }

    @Test
    public void testAddedSeldonMLDeploymentsForSingleNamespaceCluster_FailureCase() throws Exception {
        //this json does not match the proto definition for SeldonDeployment and will fail
        // that's what we want for this test
        //we expect to see a warning message in the logs for failure to parse SeldonDeployment
        createMocks("src/test/resources/deployment_model_deployed.json", "ADDED",true);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        watcher.watchSeldonMLDeployments(0,0);
        verify(mockCRDHandler,times(1)).updateRaw(anyString(),anyString(),anyString());
    }

    @Test
    public void testAddedSeldonMLDeploymentsForMultiNamespaceCluster_SuccessCase() throws Exception {
        createMocks("src/test/resources/model_simple.json", "ADDED",false);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        Assert.assertEquals("Expected and Actual resource versions do not match !!!",
                1, watcher.watchSeldonMLDeployments(0, 0));
        verify(mockSeldonDeploymentController,times(1)).createOrReplaceSeldonDeployment(any(DeploymentProtos.SeldonDeployment.class));

    }

    @Test
    public void testModifiedSeldonMLDeploymentsForMultiNamespaceCluster() throws Exception {
        createMocks("src/test/resources/model_simple.json", "MODIFIED",false);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        watcher.watchSeldonMLDeployments(0,0);
        verify(mockSeldonDeploymentController,times(1)).createOrReplaceSeldonDeployment(any(DeploymentProtos.SeldonDeployment.class));

    }

    @Test
    public void testDeletedSeldonMLDeploymentsForMultiNameSpaceCluster() throws Exception {
        createMocks("src/test/resources/model_simple.json", "DELETED",false);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        watcher.watchSeldonMLDeployments(0,0);
        verify(mockMLCache,times(1)).remove(any(DeploymentProtos.SeldonDeployment.class));
    }

    @Test
    public void testAddedSeldonMLDeploymentsForMultiNamespaceCluster_FailureCase() throws Exception {
        //this json does not match the proto definition for SeldonDeployment and will fail
        //that's what we want for this test
        //we expect to see a warning message in the logs for failure to parse SeldonDeployment
        createMocks("src/test/resources/deployment_model_deployed.json", "ADDED",false);
        SeldonDeploymentWatcher watcher = new SeldonDeploymentWatcher(mockK8sApiProvider, mockK8sClientProvider, mockCRDCreator, props, mockSeldonDeploymentController, mockMLCache, mockCRDHandler);
        watcher.watchSeldonMLDeployments(0,0);
        verify(mockCRDHandler,times(1)).updateRaw(anyString(),anyString(),anyString());
    }


}