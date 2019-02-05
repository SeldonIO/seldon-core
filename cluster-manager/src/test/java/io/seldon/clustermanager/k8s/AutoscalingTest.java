package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;

import io.kubernetes.client.proto.V2beta1Autoscaling.HorizontalPodAutoscaler;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class AutoscalingTest extends AppTest {
	
	@Test
    public void checkAutoscalerExists() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_hpa.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        DeploymentResources resources = op.createResources(mlDep2);
        
        Assert.assertEquals(1, resources.hpas.size());
        HorizontalPodAutoscaler hpa = resources.hpas.get(0);
        Assert.assertEquals("my-dep", hpa.getSpec().getScaleTargetRef().getName());
        Assert.assertEquals("Deployment", hpa.getSpec().getScaleTargetRef().getKind());
        Assert.assertEquals(SeldonDeploymentControllerImpl.DEPLOYMENT_API_VERSION, hpa.getSpec().getScaleTargetRef().getApiVersion());
    }
	
	@Test
    public void autoscalerFixDeploymentKindAndVersionTest() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_hpa_badversion.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        DeploymentResources resources = op.createResources(mlDep2);
        
        Assert.assertEquals(1, resources.hpas.size());
        HorizontalPodAutoscaler hpa = resources.hpas.get(0);
        Assert.assertEquals("my-dep", hpa.getSpec().getScaleTargetRef().getName());
        Assert.assertEquals("Deployment", hpa.getSpec().getScaleTargetRef().getKind());
        Assert.assertEquals(SeldonDeploymentControllerImpl.DEPLOYMENT_API_VERSION, hpa.getSpec().getScaleTargetRef().getApiVersion());
    }
	
	@Test(expected = SeldonDeploymentException.class)
    public void autoscalerBadTargetRefTest() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_hpa_badname.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        op.validate(mlDep2);
    }
}
