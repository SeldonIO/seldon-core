package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.seldon.clustermanager.k8s.MLDeploymentOperatorImpl.DeploymentResources;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.protos.DeploymentProtos.MLDeployment;

@Component
public class MLDeploymentControllerImpl implements MLDeploymentController {

	private final static Logger logger = LoggerFactory.getLogger(MLDeploymentControllerImpl.class);
	
	private final MLDeploymentOperator operator;
	private final K8sClientProvider clientProvider;
	
	@Autowired
	public MLDeploymentControllerImpl(MLDeploymentOperator operator, K8sClientProvider clientProvider) {
		super();
		this.operator = operator;
		this.clientProvider = clientProvider;
	}


	private void createDeployments(ApiClient client,List<ExtensionsV1beta1Deployment> deployments) throws ApiException
	{
		ExtensionsV1beta1Api api = new ExtensionsV1beta1Api(client);
		for (ExtensionsV1beta1Deployment d : deployments)
		{
			try
			{
				api.readNamespacedDeployment(d.getMetadata().getName(), "default", null, null, null);
				api.replaceNamespacedDeployment(d.getMetadata().getName(), "default", d, null);
			} 
			catch (ApiException e) 
			{
				if (e.getCode() == 404)//Not Found
				{
					api.createNamespacedDeployment("default", d, null);	
				}
				else
					throw e;
			}
		}
	}



	@Override
	public void createOrReplaceMLDeployment(MLDeployment mlDep) {
		
		try
		{
			mlDep = operator.defaulting(mlDep);
			operator.validate(mlDep);
			DeploymentResources resources = operator.createResources(mlDep);
			ApiClient client = clientProvider.getClient();
			createDeployments(client, resources.deployments);
			
		} catch (MLDeploymentException e) {
			logger.error("Failed to create deployment ",e);
		} catch (ApiException e) {
			//FIXME Parse response body to get enclosing message?
			logger.error("Kubernetes API exception deploying code:"+e.getCode()+ "message:"+e.getResponseBody(),e);
		} catch (IOException e) {
			logger.error("Failed to get API Client ",e);
		}
		finally {}
		
	}

}
