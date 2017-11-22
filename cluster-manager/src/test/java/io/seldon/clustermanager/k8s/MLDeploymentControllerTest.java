package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.client.K8sDefaultClientProvider;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public class MLDeploymentControllerTest extends AppTest {
	
	@Test
	public void simpleTest() throws IOException
	{
		MLDeploymentOperator op = new MLDeploymentOperatorImpl(getProps());
		MLDeploymentController controller = new MLDeploymentControllerImpl(op, new K8sDefaultClientProvider());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		MLDeployment mlDep = MLDeploymentUtils.jsonToMLDeployment(jsonStr);
		controller.createOrReplaceMLDeployment(mlDep);
	}
	
	
}
