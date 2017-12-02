package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.client.K8sDefaultClientProvider;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeploymentControllerExample extends AppTest {
	
    public void run() throws IOException
	{
		SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getProps());
		KubeCRDHandler crdHandler = new KubeCRDHandlerImpl();
		SeldonDeploymentController controller = new SeldonDeploymentControllerImpl(op, new K8sDefaultClientProvider(), crdHandler, new SeldonDeploymentCacheImpl(crdHandler));
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		controller.createOrReplaceSeldonDeployment(mlDep);
	}
	
    public static void main(String[] args) throws IOException {
        SeldonDeploymentControllerExample e = new SeldonDeploymentControllerExample();
        e.run();
    }
}
