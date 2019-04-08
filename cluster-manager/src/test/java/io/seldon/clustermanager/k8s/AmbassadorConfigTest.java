package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Test;

import io.kubernetes.client.proto.V1.Service;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class AmbassadorConfigTest extends AppTest {
		
	    @Test
	    public void checkAmbassadorConfigExists() throws IOException, SeldonDeploymentException
	    {
	        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
	        String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
	        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
	        SeldonDeployment mlDep2 = op.defaulting(mlDep);
	        DeploymentResources resources = op.createResources(mlDep2);
	        
	        Service service = resources.services.get(resources.services.size()-1);
	        String ambassadorConfig = service.getMetadata().getAnnotationsOrDefault("getambassador.io/config", null);
	        System.out.println(ambassadorConfig);	        
	        Assert.assertNotNull(ambassadorConfig);
	    }
	    
	    @Test
	    public void checkAmbassadorCanary() throws IOException, SeldonDeploymentException
	    {
	        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
	        String jsonStr = readFile("src/test/resources/model_ambassador_canary.json",StandardCharsets.UTF_8);
	        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
	        SeldonDeployment mlDep2 = op.defaulting(mlDep);
	        DeploymentResources resources = op.createResources(mlDep2);
	        
	        Service service = resources.services.get(resources.services.size()-1);
	        String ambassadorConfig = service.getMetadata().getAnnotationsOrDefault("getambassador.io/config", null);
	        Assert.assertNotNull(ambassadorConfig);
	        System.out.println(ambassadorConfig);
	        Assert.assertTrue(ambassadorConfig.indexOf("weight: 25\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("prefix: /seldon/default/example/\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("apiVersion: ambassador/v1")==4);
	    }
	    
	    @Test
	    public void checkAmbassadorShadow() throws IOException, SeldonDeploymentException
	    {
	        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
	        String jsonStr = readFile("src/test/resources/model_ambassador_shadow.json",StandardCharsets.UTF_8);
	        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
	        SeldonDeployment mlDep2 = op.defaulting(mlDep);
	        DeploymentResources resources = op.createResources(mlDep2);
	        
	        Service service = resources.services.get(resources.services.size()-1);
	        String ambassadorConfig = service.getMetadata().getAnnotationsOrDefault("getambassador.io/config", null);
	        Assert.assertNotNull(ambassadorConfig);
	        System.out.println(ambassadorConfig);
	        Assert.assertTrue(ambassadorConfig.indexOf("shadow: true\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("prefix: /seldon/default/example/\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("apiVersion: ambassador/v1")==4);
	    }
	    
	    @Test
	    public void checkAmbassadorHeader() throws IOException, SeldonDeploymentException
	    {
	        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
	        String jsonStr = readFile("src/test/resources/model_ambassador_headers.json",StandardCharsets.UTF_8);
	        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
	        SeldonDeployment mlDep2 = op.defaulting(mlDep);
	        DeploymentResources resources = op.createResources(mlDep2);
	        
	        Service service = resources.services.get(resources.services.size()-1);
	        String ambassadorConfig = service.getMetadata().getAnnotationsOrDefault("getambassador.io/config", null);
	        Assert.assertNotNull(ambassadorConfig);
	        System.out.println(ambassadorConfig);
	        Assert.assertTrue(ambassadorConfig.indexOf("  location: london\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("prefix: /seldon/default/example/\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("apiVersion: ambassador/v1")==4);
	    }	
	    
	    @Test
	    public void checkAmbassadorCustomConfig() throws IOException, SeldonDeploymentException
	    {
	        SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getClusterManagerprops());
	        String jsonStr = readFile("src/test/resources/model_ambassador_config.json",StandardCharsets.UTF_8);
	        SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
	        SeldonDeployment mlDep2 = op.defaulting(mlDep);
	        DeploymentResources resources = op.createResources(mlDep2);
	        
	        Service service = resources.services.get(resources.services.size()-1);
	        String ambassadorConfig = service.getMetadata().getAnnotationsOrDefault("getambassador.io/config", null);
	        Assert.assertNotNull(ambassadorConfig);
	        System.out.println(ambassadorConfig);
	        Assert.assertTrue(ambassadorConfig.indexOf("1234")==0);
	        Assert.assertFalse(ambassadorConfig.indexOf("prefix: /seldon/default/example/\n")>0);
	        Assert.assertTrue(ambassadorConfig.indexOf("apiVersion: ambassador/v1")==-1);
	    }	    
	    
}
