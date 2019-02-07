package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class DuplicateComponentGraphTest  extends AppTest {
	
    @Test
    public void duplicateComponentTest() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_duplicate_component.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        op.validate(mlDep2);
        DeploymentResources resources = op.createResources(mlDep2);
        
        Assert.assertEquals(1, resources.deployments.size());
        Assert.assertEquals(2, resources.deployments.get(0).getSpec().getTemplate().getSpec().getContainersCount());
    }
    

}
