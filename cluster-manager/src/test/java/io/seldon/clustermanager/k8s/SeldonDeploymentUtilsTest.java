package io.seldon.clustermanager.k8s;

import org.junit.Test;

import org.junit.Assert;

public class SeldonDeploymentUtilsTest {

	
	@Test
	public void getVersionTest()
	{
		String version = SeldonDeploymentUtils.getVersionFromApiVersion("machinelearning.seldon.io/v1alpha2");
		Assert.assertEquals("v1alpha2", version);
	}
	
	@Test(expected = ArrayIndexOutOfBoundsException.class)
	public void getVersionTestFailed()
	{
		String version = SeldonDeploymentUtils.getVersionFromApiVersion("machinelearning.seldon.io.v1alpha2");
		Assert.assertEquals("v1alpha2", version);
	}
	
}
