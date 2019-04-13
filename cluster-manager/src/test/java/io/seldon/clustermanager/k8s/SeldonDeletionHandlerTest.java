package io.seldon.clustermanager.k8s;

import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.never;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import org.junit.Test;

import com.google.protobuf.Message;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1ObjectMeta;
import io.kubernetes.client.proto.Meta.DeleteOptions;
import io.kubernetes.client.proto.V2beta1Autoscaling.HorizontalPodAutoscaler;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeletionHandlerTest   extends AppTest {

	
	
	@Test
	public void testDeploymentDeleted() throws IOException, ApiException, SeldonDeploymentException
	{
		ClusterManagerProperites props = new ClusterManagerProperites();
		props.setEngineContainerImageAndVersion("engine:0.1");
		props.setEngineContainerImagePullPolicy("IfNotPresent");
		props.setEngineContainerServiceAccountName("default");
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeploymentOperatorImpl operator = new SeldonDeploymentOperatorImpl(props);
        String namespace = SeldonDeploymentUtils.getNamespace(mlDep);
        DeploymentResources resources = operator.createResources(mlDep);
        
		KubeCRDHandler crdHandler= mock(KubeCRDHandler.class);
		ApiClient apiClient = new ApiClient();
		
		ExtensionsV1beta1Deployment d = new ExtensionsV1beta1Deployment();
		V1ObjectMeta d1Meta = new V1ObjectMeta();
		d1Meta.setName("XYZ");
		d.setMetadata(d1Meta);
		List<ExtensionsV1beta1Deployment> dList = new ArrayList<>();
		dList.add(d);
		ExtensionsV1beta1DeploymentList deployments = new ExtensionsV1beta1DeploymentList();
		deployments.setItems(dList);
		
		when(crdHandler.getOwnedDeployments(any(String.class), any(String.class))).thenReturn(deployments);
		ProtoClient client = mock(ProtoClient.class);
		when(client.getApiClient()).thenReturn(apiClient);
		when(client.delete(any(Message.Builder.class), any(String.class), any(DeleteOptions.class))).thenReturn(new ObjectOrStatus<>(mlDep,null));
		
        
        SeldonDeletionHandler delHandler = new SeldonDeletionHandler(crdHandler);
        
        delHandler.removeDeployments(client, namespace, mlDep, resources.deployments, false);
        verify(client).delete(any(Message.Builder.class), any(String.class), any(DeleteOptions.class));
        
	}
	
	
	@Test
	public void testDeploymentsNoDelete() throws IOException, ApiException, SeldonDeploymentException
	{
		ClusterManagerProperites props = new ClusterManagerProperites();
		props.setEngineContainerImageAndVersion("engine:0.1");
		props.setEngineContainerImagePullPolicy("IfNotPresent");
		props.setEngineContainerServiceAccountName("default");
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeploymentOperatorImpl operator = new SeldonDeploymentOperatorImpl(props);
        String namespace = SeldonDeploymentUtils.getNamespace(mlDep);
        DeploymentResources resources = operator.createResources(mlDep);
        
		KubeCRDHandler crdHandler= mock(KubeCRDHandler.class);
		ApiClient apiClient = new ApiClient();
		
		ExtensionsV1beta1Deployment d = new ExtensionsV1beta1Deployment();
		V1ObjectMeta d1Meta = new V1ObjectMeta();
		d1Meta.setName(resources.deployments.get(0).getMetadata().getName());
		d.setMetadata(d1Meta);
		List<ExtensionsV1beta1Deployment> dList = new ArrayList<>();
		dList.add(d);
		ExtensionsV1beta1DeploymentList deployments = new ExtensionsV1beta1DeploymentList();
		deployments.setItems(dList);
		
		when(crdHandler.getOwnedDeployments(any(String.class), any(String.class))).thenReturn(deployments);
		ProtoClient client = mock(ProtoClient.class);
		when(client.getApiClient()).thenReturn(apiClient);
		when(client.delete(any(Message.Builder.class), any(String.class), any(DeleteOptions.class))).thenReturn(new ObjectOrStatus<>(mlDep,null));
		
		when(crdHandler.getOwnedHPAs(any(String.class), any(String.class))).thenReturn(Optional.empty());
        
        SeldonDeletionHandler delHandler = new SeldonDeletionHandler(crdHandler);
        
        delHandler.removeDeployments(client, namespace, mlDep, resources.deployments, false);
        verify(client, never()).delete(any(Message.Builder.class), any(String.class), any(DeleteOptions.class));
        
        List<HorizontalPodAutoscaler> hpas = new ArrayList<>();
        delHandler.removeHPAs(apiClient, namespace, mlDep, hpas);
        
	}
	
}
