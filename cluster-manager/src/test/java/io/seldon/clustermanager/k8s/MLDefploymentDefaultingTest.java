package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.clustermanager.AppTest;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public class MLDefploymentDefaultingTest extends AppTest {
	private final static Logger logger = LoggerFactory.getLogger(MLDefploymentDefaultingTest.class);
	
	@Test
	public void testDefaulting() throws IOException
	{
		MLDeploymentOperator op = new MLDeploymentOperatorImpl(getProps());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		MLDeployment mlDep = MLDeploymentUtils.jsonToMLDeployment(jsonStr);
		MLDeployment mlDep2 = op.defaulting(mlDep);
		logger.info(MLDeploymentUtils.toJson(mlDep2));
	}
}
