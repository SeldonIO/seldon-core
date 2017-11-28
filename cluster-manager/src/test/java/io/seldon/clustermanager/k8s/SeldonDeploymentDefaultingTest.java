package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.clustermanager.AppTest;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeploymentDefaultingTest extends AppTest {
	private final static Logger logger = LoggerFactory.getLogger(SeldonDeploymentDefaultingTest.class);
	
	@Test
	public void testDefaulting() throws IOException
	{
		SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getProps());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		SeldonDeployment mlDep2 = op.defaulting(mlDep);
		logger.info(SeldonDeploymentUtils.toJson(mlDep2,false));
	}
}
