/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.clustermanager.AppTest;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeploymentDefaultingTest extends AppTest {
	
    @Test
    public void testDefaulting() throws IOException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        
        Assert.assertTrue(mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).hasLivenessProbe());
        Assert.assertTrue(mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).hasReadinessProbe());
        Assert.assertTrue(mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).hasLifecycle());
        Assert.assertEquals("Incorrect number of environment variables in container",5,mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).getEnvCount());
        Assert.assertEquals(1,mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).getPortsCount());
        Assert.assertEquals("http",mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).getPorts(0).getName());
        Assert.assertEquals(Endpoint.EndpointType.REST_VALUE,mlDep2.getSpec().getPredictors(0).getGraph().getEndpoint().getType().getNumber());
        Assert.assertEquals("0.0.0.0",mlDep2.getSpec().getPredictors(0).getGraph().getEndpoint().getServiceHost());
    }

    @Test
    public void testDefaultingGrpc() throws IOException
    {
        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
        String jsonStr = readFile("src/test/resources/model_simple_grpc.json",StandardCharsets.UTF_8);
        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        SeldonDeployment mlDep2 = op.defaulting(mlDep);
        
        Assert.assertTrue(mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).hasLivenessProbe());
        Assert.assertTrue(mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).hasReadinessProbe());
        Assert.assertTrue(mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).hasLifecycle());
        Assert.assertEquals("Incorrect number of environment variables in container",5,mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).getEnvCount());
        Assert.assertEquals(1,mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).getPortsCount());
        Assert.assertEquals("grpc",mlDep2.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0).getPorts(0).getName());
        Assert.assertEquals(Endpoint.EndpointType.GRPC_VALUE,mlDep2.getSpec().getPredictors(0).getGraph().getEndpoint().getType().getNumber());
        Assert.assertEquals("0.0.0.0",mlDep2.getSpec().getPredictors(0).getGraph().getEndpoint().getServiceHost());
    }
}
