package io.seldon.apife.deployments;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Test;

import io.seldon.apife.AppProperties;
import io.seldon.apife.SeldonTestBase;
import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;
import org.junit.Assert;

public class DeploymentStoreTest extends SeldonTestBase {

	@Test
	public void deploymentSingleNamespaceTest() throws IOException
	{
		InMemoryClientDetailsService clientDetailsService = new InMemoryClientDetailsService();
		AppProperties appProperties = new AppProperties();
		DeploymentStore deploymentStore = new DeploymentStore(new DeploymentsHandler() {
			
			@Override
			public void addListener(DeploymentsListener listener) {
				// TODO Auto-generated method stub
				
			}
		}, clientDetailsService, appProperties);
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
		SeldonDeployment.Builder dBuilder = SeldonDeployment.newBuilder();
		updateMessageBuilderFromJson(dBuilder, jsonStr);
		SeldonDeployment mlDep = dBuilder.build();
		
		deploymentStore.deploymentAdded(mlDep);
		
		Assert.assertNotNull(deploymentStore.getDeployment(mlDep.getSpec().getOauthKey()));
		Assert.assertNotNull(deploymentStore.getDeployment(mlDep.getSpec().getOauthKey()+mlDep.getMetadata().getNamespace()));
	}
	
	@Test
	public void deploymentClusterWideTest() throws IOException
	{
		InMemoryClientDetailsService clientDetailsService = new InMemoryClientDetailsService();
		AppProperties appProperties = new AppProperties();
		appProperties.setSingleNamespace(false);
		DeploymentStore deploymentStore = new DeploymentStore(new DeploymentsHandler() {
			
			@Override
			public void addListener(DeploymentsListener listener) {
				// TODO Auto-generated method stub
				
			}
		}, clientDetailsService, appProperties);
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
		SeldonDeployment.Builder dBuilder = SeldonDeployment.newBuilder();
		updateMessageBuilderFromJson(dBuilder, jsonStr);
		SeldonDeployment mlDep = dBuilder.build();
		
		deploymentStore.deploymentAdded(mlDep);
		
		Assert.assertNull(deploymentStore.getDeployment(mlDep.getSpec().getOauthKey()));
		Assert.assertNotNull(deploymentStore.getDeployment(mlDep.getSpec().getOauthKey()+mlDep.getMetadata().getNamespace()));
	}
	
}
