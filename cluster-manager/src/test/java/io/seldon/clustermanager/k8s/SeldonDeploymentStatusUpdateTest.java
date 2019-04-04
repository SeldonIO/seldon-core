package io.seldon.clustermanager.k8s;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;
import static org.mockito.Mockito.never;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;
import org.mockito.ArgumentCaptor;

import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;


public class SeldonDeploymentStatusUpdateTest extends AppTest {
	
	KubeCRDHandler mockCrdHandler;
	SeldonDeploymentController mockSeldonDeploymentController;
	ClusterManagerProperites props;
	
	private void createMocks()
	{
		mockCrdHandler = mock(KubeCRDHandler.class);
		mockSeldonDeploymentController = mock(SeldonDeploymentController.class);
		props = new ClusterManagerProperites();
		props.setEngineContainerImageAndVersion("engine:0.1");
		props.setEngineContainerImagePullPolicy("IfNotPresent");
		props.setEngineContainerServiceAccountName("default");
	}
	
	@Test
	public void updateAvailableTest() throws IOException
	{
		createMocks();
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment sDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		
		when(mockCrdHandler.getSeldonDeployment(any(String.class), any(String.class), any(String.class))).thenReturn(sDep);
		
		SeldonDeploymentStatusUpdate supdate = new SeldonDeploymentStatusUpdateImpl(mockCrdHandler, mockSeldonDeploymentController, props);
		
		final String selDepName = "SeldonDep1";
		final String version = "v1alpha1";
		final String namespace = "seldon";
		
		supdate.updateStatus(selDepName, version, "test-deployment-fx-market-predictor-8e1d76f", 1, 1, namespace);
		
		verify(mockCrdHandler,times(1)).getSeldonDeployment(eq(selDepName), eq(version), eq(namespace));
		verify(mockCrdHandler,times(1)).updateSeldonDeploymentStatus(any(SeldonDeployment.class));

	}
	
	@Test
	public void updateAvailableOnFailedTest() throws IOException
	{
		createMocks();
		
		String jsonStr = readFile("src/test/resources/model_failed.json",StandardCharsets.UTF_8);
        SeldonDeployment sDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		
		when(mockCrdHandler.getSeldonDeployment(any(String.class), any(String.class), any(String.class))).thenReturn(sDep);
		
		SeldonDeploymentStatusUpdate supdate = new SeldonDeploymentStatusUpdateImpl(mockCrdHandler, mockSeldonDeploymentController, props);
		
		final String selDepName = "SeldonDep1";
		final String version = "v1alpha1";
		final String namespace = "seldon";
		
		supdate.updateStatus(selDepName, version, "test-deployment-fx-market-predictor-8e1d76f", 1, 1, namespace);
		
		verify(mockCrdHandler,times(1)).getSeldonDeployment(eq(selDepName), eq(version), eq(namespace));
		ArgumentCaptor<SeldonDeployment> argument = ArgumentCaptor.forClass(SeldonDeployment.class);
		verify(mockCrdHandler,never()).updateSeldonDeploymentStatus(argument.capture());
		
		SeldonDeployment sDepUpdated = argument.getAllValues().get(0);
		Assert.assertEquals(1, sDepUpdated.getStatus().getPredictorStatusCount());
		Assert.assertEquals(1, sDepUpdated.getStatus().getPredictorStatus(0).getReplicasAvailable());
		Assert.assertEquals("Available", sDepUpdated.getStatus().getState());
	}
	
	
	@Test
	public void twoUpdatesTest() throws IOException
	{
		createMocks();
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment sDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		
		when(mockCrdHandler.getSeldonDeployment(any(String.class), any(String.class), any(String.class))).thenReturn(sDep);
		
		SeldonDeploymentStatusUpdate supdate = new SeldonDeploymentStatusUpdateImpl(mockCrdHandler, mockSeldonDeploymentController, props);
		
		final String selDepName = "SeldonDep1";
		final String version = "v1alpha1";
		final String namespace = "seldon";
		
		supdate.updateStatus(selDepName, version, "test-deployment-fx-market-predictor-8e1d76f", 1, 0, namespace);
		
		verify(mockCrdHandler,times(1)).getSeldonDeployment(eq(selDepName), eq(version), eq(namespace));
		ArgumentCaptor<SeldonDeployment> argument = ArgumentCaptor.forClass(SeldonDeployment.class);
		verify(mockCrdHandler,times(1)).updateSeldonDeploymentStatus(argument.capture());
		
		SeldonDeployment sDepUpdated = argument.getAllValues().get(0);
		Assert.assertEquals(1, sDepUpdated.getStatus().getPredictorStatusCount());
		Assert.assertEquals(0, sDepUpdated.getStatus().getPredictorStatus(0).getReplicasAvailable());
		Assert.assertEquals("Creating", sDepUpdated.getStatus().getState());
		
		supdate.updateStatus(selDepName, version, "test-deployment-fx-market-predictor-8e1d76f", 1, 1, namespace);
		
		verify(mockCrdHandler,times(2)).getSeldonDeployment(eq(selDepName), eq(version), eq(namespace));
		argument = ArgumentCaptor.forClass(SeldonDeployment.class);
		verify(mockCrdHandler,times(2)).updateSeldonDeploymentStatus(argument.capture());
		sDepUpdated = argument.getAllValues().get(1);
		Assert.assertEquals(1, sDepUpdated.getStatus().getPredictorStatusCount());
		Assert.assertEquals(1, sDepUpdated.getStatus().getPredictorStatus(0).getReplicasAvailable());
		Assert.assertEquals("Available", sDepUpdated.getStatus().getState());
	}
	
	@Test
	public void removeTest() throws IOException
	{
		createMocks();
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment sDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		
		when(mockCrdHandler.getSeldonDeployment(any(String.class), any(String.class), any(String.class))).thenReturn(sDep);
		
		SeldonDeploymentStatusUpdate supdate = new SeldonDeploymentStatusUpdateImpl(mockCrdHandler, mockSeldonDeploymentController, props);
		
		final String selDepName = "SeldonDep1";
		final String version = "v1alpha1";
		final String namespace = "seldon";
		
		supdate.removeStatus(selDepName, version, "test-deployment-fx-market-predictor-8e1d76f", namespace);
		
		ArgumentCaptor<SeldonDeployment> argument = ArgumentCaptor.forClass(SeldonDeployment.class);
		verify(mockCrdHandler,times(1)).updateSeldonDeploymentStatus(argument.capture());
		SeldonDeployment sDepUpdated = argument.getAllValues().get(0);
		Assert.assertEquals(0, sDepUpdated.getStatus().getPredictorStatusCount());
		Assert.assertEquals("", sDepUpdated.getStatus().getState());
	}
	
	
	

}
