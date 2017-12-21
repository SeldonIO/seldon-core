package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Test;

import io.seldon.clustermanager.AppTest;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeploymentValidationTest extends AppTest {
    @Test
    public void testDefaulting() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        op.validate(mlDep2);
    }
    
    @Test(expected = SeldonDeploymentException.class)
    public void testBadGraph() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_invalid_graph.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        op.validate(mlDep2);
    }

    @Test(expected = SeldonDeploymentException.class)
    public void testNoType() throws IOException, SeldonDeploymentException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_invalid_no_type.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        op.validate(mlDep2);
    }

}
