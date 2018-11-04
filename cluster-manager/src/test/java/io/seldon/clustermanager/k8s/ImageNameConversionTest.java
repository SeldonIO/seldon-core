package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;

import io.kubernetes.client.proto.V1;
import io.seldon.clustermanager.AppTest;
import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class ImageNameConversionTest extends AppTest {

	@Test
	public void testImageName()
	{
		String imageName = "seldonio/MEAN-classifier***:0.1.2";
		Assert.assertFalse(imageName.matches("[a-z]([-a-z0-9]*[a-z0-9])?"));
		String changed = SeldonNameCreator.cleanContainerImageName(imageName);
		Assert.assertTrue(changed.matches("[a-z]([-a-z0-9]*[a-z0-9])?"));
	}
	
	@Test 
	public void checkServiceName_short() throws IOException
	{
		SeldonNameCreator seldonNameCreator = new SeldonNameCreator();
		final String containerImage = "seldonio/mean-classifier:0.1.2";
		final String containerName = "c";
		String jsonStr = readFile("src/test/resources/model_short_names.json",StandardCharsets.UTF_8);
        SeldonDeployment dep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        V1.Container c =dep.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0);
		String name = seldonNameCreator.getSeldonServiceName(dep, dep.getSpec().getPredictors(0), c);
		System.out.println(name);
		Assert.assertTrue(name.matches("[a-z]([-a-z0-9]*[a-z0-9])?"));//Valid DNS name
	}

	@Test 
	public void checkServiceName_medium() throws IOException
	{
		SeldonNameCreator seldonNameCreator = new SeldonNameCreator();
		final String containerImage = "seldonio/mean-classifier:0.1.2";
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
        SeldonDeployment dep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        V1.Container c =dep.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0);
		String name = seldonNameCreator.getSeldonServiceName(dep, dep.getSpec().getPredictors(0),c);
		System.out.println(name);
		Assert.assertTrue(name.matches("[a-z]([-a-z0-9]*[a-z0-9])?"));//Valid DNS name
	}

	@Test 
	public void checkServiceName_long() throws IOException
	{
		SeldonNameCreator seldonNameCreator = new SeldonNameCreator();
		final String containerImage = "seldonio/mean-classifier-------------------------------long:0.1.2";
		String jsonStr = readFile("src/test/resources/model_long_names.json",StandardCharsets.UTF_8);
        SeldonDeployment dep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
        V1.Container c =dep.getSpec().getPredictors(0).getComponentSpecs(0).getSpec().getContainers(0);
		String name = seldonNameCreator.getSeldonServiceName(dep, dep.getSpec().getPredictors(0), c);
		System.out.println(name);
		Assert.assertTrue(name.matches("[a-z]([-a-z0-9]*[a-z0-9])?"));//Valid DNS name
	}
	
}
