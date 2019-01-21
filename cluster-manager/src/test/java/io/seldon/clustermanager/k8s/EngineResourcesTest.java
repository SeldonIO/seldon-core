package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;

import io.kubernetes.client.proto.V1.Container;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class EngineResourcesTest extends AppTest {
	
    @Test
    public void testDefaultEngineResources() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        DeploymentResources resources = op.createResources(mlDep2);
        
        Container engineContainer = resources.deployments.get(0).getSpec().getTemplate().getSpec().getContainers(1);
        Assert.assertEquals("seldon-container-engine", engineContainer.getName());
        Assert.assertEquals(SeldonDeploymentOperatorImpl.DEFAULT_ENGINE_CPU_REQUEST, engineContainer.getResources().getRequestsOrThrow("cpu").getString());
    }
    
    @Test
    public void testEngineSvcOrchResources() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_svcorch_resources.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        DeploymentResources resources = op.createResources(mlDep2);
        
        Container engineContainer = resources.deployments.get(0).getSpec().getTemplate().getSpec().getContainers(1);
        Assert.assertEquals("seldon-container-engine", engineContainer.getName());
        Assert.assertEquals("1Mi", engineContainer.getResources().getRequestsOrThrow("memory").getString());
    }

    @Test
    public void testEngineResources() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_engine_resources.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        DeploymentResources resources = op.createResources(mlDep2);
        
        Container engineContainer = resources.deployments.get(0).getSpec().getTemplate().getSpec().getContainers(1);
        Assert.assertEquals("seldon-container-engine", engineContainer.getName());
        Assert.assertEquals("2", engineContainer.getResources().getRequestsOrThrow("cpu").getString());
    }

}
