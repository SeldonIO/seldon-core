package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;

public class K8sResourceLoadTest   {
	private final static Logger logger = LoggerFactory.getLogger(K8sResourceLoadTest.class);
	
	final String json = "{\"kind\": \"Deployment\",\"apiVersion\": \"extensions/v1beta1\",\"metadata\": {\"name\": \"kubectl-tester\"},\"spec\": {\"template\":{\"spec\":{\"containers\": [{\"name\": \"bb\",\"image\": \"gcr.io/google_containers/busybox\"}]}}}}";
	@Test
	public void ResourceTest() throws IOException
	{
		Config config = new ConfigBuilder().build(); // use defaults for config
        logger.info(String.format("Connecting to kubernetes master[%s]", config.getMasterUrl()));
        KubernetesClient kubernetesClient = new DefaultKubernetesClient(config);

        Deployment dep = kubernetesClient.extensions().deployments().load("src/test/resources/deployment.yaml").get();

        logger.info(dep.getSpec().getTemplate().getSpec().getContainers().get(0).getImage());
 
        getNamespaceList(kubernetesClient); // simple check to see if client works
        logger.info("Sucessfully passed getNamespaceList() check");
       
	}
	 public List<String> getNamespaceList(KubernetesClient kubernetesClient) {
	        List<String> namespace_list = new ArrayList<>();
	        NamespaceList namespaceList = kubernetesClient.namespaces().list();
	        for (Namespace ns : namespaceList.getItems()) {
	            namespace_list.add(ns.getMetadata().getName());
	        }

	        return namespace_list;
	    }
}
