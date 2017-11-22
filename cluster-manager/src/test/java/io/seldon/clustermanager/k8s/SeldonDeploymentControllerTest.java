package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Test;

import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.client.K8sDefaultClientProvider;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeploymentControllerTest extends AppTest {
	
	@Test
	public void simpleTest() throws IOException
	{
		SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getProps());
		SeldonDeploymentController controller = new SeldonDeploymentControllerImpl(op, new K8sDefaultClientProvider());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToMLDeployment(jsonStr);
		controller.createOrReplaceMLDeployment(mlDep);
	}
	
	
}
