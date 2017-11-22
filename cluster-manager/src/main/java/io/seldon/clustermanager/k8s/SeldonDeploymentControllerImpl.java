package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.kubernetes.client.ApiException;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentControllerImpl implements SeldonDeploymentController {

	private final static Logger logger = LoggerFactory.getLogger(SeldonDeploymentControllerImpl.class);
	
	private final SeldonDeploymentOperator operator;
	private final K8sClientProvider clientProvider;
	
	private static final String DEPLOYMENT_API_VERSION = "extensions/v1beta1";
	
	@Autowired
	public SeldonDeploymentControllerImpl(SeldonDeploymentOperator operator, K8sClientProvider clientProvider) {
		super();
		this.operator = operator;
		this.clientProvider = clientProvider;
	}


	/*
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
	*/
	
	
	
	private void createDeployments(ProtoClient client,List<Deployment> deployments) throws ApiException, IOException, SeldonDeploymentException
	{
		for(Deployment d : deployments)
		{
		    final String listApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
	                .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(d.getMetadata().getName()))
	                .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString("default"));
		    ObjectOrStatus<Deployment> os = client.list(Deployment.newBuilder(),listApiPath);       
            if (os.status != null) {
                if (os.status.getCode() == 404) { //Create
                    logger.info("About to create "+ProtoBufUtils.toJson(d));
                    final String createApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments"
                            .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString("default"));

                    os = client.create(d, createApiPath, DEPLOYMENT_API_VERSION, "Deployment");
                    if (os.status != null) {
                        logger.error("Error creating deployment:"+ProtoBufUtils.toJson(os.status));
                        throw new SeldonDeploymentException("Failed to create deployment "+d.getMetadata().getName());
                    }
                    else {
                        logger.debug("Created deployment:"+ProtoBufUtils.toJson(os.object));
                    }
                }
                else {
                    logger.error("Error listing deployment:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to list deployment "+d.getMetadata().getName());
                }
            }
            else { // Update
                os = client.update(d,listApiPath, DEPLOYMENT_API_VERSION, "Deployment");
                if (os.status != null) {
                    logger.error("Error updating deployment:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to update deployment "+d.getMetadata().getName());
                }
                else {
                    logger.debug("Created deployment:"+ProtoBufUtils.toJson(os.object));
                }
            }
		}
	}

	private void createService(ProtoClient client,Service service) throws ApiException, IOException, SeldonDeploymentException
	{
	    final String serviceApiPath = "/api/v1/namespaces/{namespace}/services/{name}"
                .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(service.getMetadata().getName()))
                .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString("default"));
        ObjectOrStatus os = client.list(Service.newBuilder(),serviceApiPath);     
        if (os.status != null)
        {
            if (os.status.getCode() == 404)
            {
                String serviceCreateApiPath = "/api/v1/namespaces/{namespace}/services"
                        .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString("default"));
                os = client.create(service, serviceCreateApiPath, "v1", "Service");
                if (os.status != null)
                {
                    logger.error("Error creating service "+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to create service "+service.getMetadata().getName());
                }
                else
                {
                    logger.debug("Created service:"+ProtoBufUtils.toJson(os.object));
                }                   
            }
            else
            {
                logger.error("Error listing service:"+ProtoBufUtils.toJson(os.status));
                throw new SeldonDeploymentException("Failed to list service "+service.getMetadata().getName());
            }
        }
        else
            logger.debug("No creating service as already exists "+service.getMetadata().getName());
	}

	@Override
	public void createOrReplaceMLDeployment(SeldonDeployment mlDep) {
		
		try
		{
			mlDep = operator.defaulting(mlDep);
			operator.validate(mlDep);
			logger.info(ProtoBufUtils.toJson(mlDep));
			DeploymentResources resources = operator.createResources(mlDep);
			ProtoClient client = clientProvider.getProtoClient();
			createDeployments(client, resources.deployments);
			createService(client,resources.service);
			
		} catch (SeldonDeploymentException e) {
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
