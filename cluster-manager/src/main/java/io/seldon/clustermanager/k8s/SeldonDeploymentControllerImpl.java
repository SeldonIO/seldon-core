package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.lang.reflect.Type;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.gson.reflect.TypeToken;

import io.kubernetes.client.ApiException;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.auth.ApiKeyAuth;
import io.kubernetes.client.auth.Authentication;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1PodTemplateSpec;
import io.kubernetes.client.proto.Meta.DeleteOptions;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentControllerImpl implements SeldonDeploymentController {

	private final static Logger logger = LoggerFactory.getLogger(SeldonDeploymentControllerImpl.class);
	private final static String SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY = "SELDON_CLUSTER_MANAGER_POD_NAMESPACE";
	private final SeldonDeploymentOperator operator;
	private final K8sClientProvider clientProvider;
	private final KubeCRDHandler crdHandler;
	private final SeldonDeploymentCache mlCache;
	
	private static final String DEPLOYMENT_API_VERSION = "extensions/v1beta1";
	private String seldonClusterNamespaceName = "UNKOWN_NAMESPACE";

	@Autowired
	public SeldonDeploymentControllerImpl(SeldonDeploymentOperator operator, K8sClientProvider clientProvider,KubeCRDHandler crdHandler,SeldonDeploymentCache mlCache) {
		super();
		this.operator = operator;
		this.clientProvider = clientProvider;
		this.crdHandler = crdHandler;
		this.mlCache = mlCache;
		
		 { // set the namespace to use
	            seldonClusterNamespaceName = System.getenv().get(SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY);
	            if (seldonClusterNamespaceName == null) {
	                logger.info(String.format("FAILED to find env var [%s]", SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY));
	                seldonClusterNamespaceName = "default";
	            }
	            logger.info(String.format("Setting cluster manager namespace as [%s]", seldonClusterNamespaceName));
	        }
	}
	
	private void createDeployments(ProtoClient client,List<Deployment> deployments) throws ApiException, IOException, SeldonDeploymentException
	{
		for(Deployment d : deployments)
		{
            final String listApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(d.getMetadata().getName()))
                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString("default"));
            logger.debug("Will try to call LIST "+listApiPath);
            ObjectOrStatus<Deployment> os = client.list(Deployment.newBuilder(),listApiPath);       
            if (os.status != null) {
                if (os.status.getCode() == 404) { //Create
                    logger.debug("About to CREATE "+ProtoBufUtils.toJson(d));
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
                logger.debug("About to UPDATE "+ProtoBufUtils.toJson(d));
                os = client.update(d,listApiPath, DEPLOYMENT_API_VERSION, "Deployment");
                if (os.status != null) {
                    logger.error("Error updating deployment:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to update deployment "+d.getMetadata().getName());
                }
                else {
                    logger.debug("Updated deployment:"+ProtoBufUtils.toJson(os.object));
                }
            }
		}
	}

	private Set<String> getDeploymentNames(List<Deployment> deployments)
	{
	    Set<String> names = new HashSet<>();
	    for(Deployment d : deployments)
	        names.add(d.getMetadata().getName());
	    return names;
	}
	
	private void removeDeployments(ProtoClient client,SeldonDeployment seldonDeployment,List<Deployment> deployments) throws ApiException, IOException, SeldonDeploymentException
	{
	    Set<String> names = getDeploymentNames(deployments);
	    ExtensionsV1beta1DeploymentList depList = crdHandler.getOwnedDeployments(seldonDeployment.getSpec().getName());
	    for (ExtensionsV1beta1Deployment d : depList.getItems())
	    {
	        if (!names.contains(d.getMetadata().getName()))
	        {
	            final String deleteApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
	                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(d.getMetadata().getName()))
	                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString("default"));
	            DeleteOptions options = DeleteOptions.newBuilder().setPropagationPolicy("Foreground").build();
	            ObjectOrStatus<Deployment> os = client.delete(Deployment.newBuilder(),deleteApiPath,options);
	            if (os.status != null) {
                    logger.error("Error deleting deployment:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to delete deployment "+d.getMetadata().getName());
                }
                else {
                    logger.debug("Deleted deployment:"+ProtoBufUtils.toJson(os.object));
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
	public void createOrReplaceSeldonDeployment(SeldonDeployment mlDep) {
		
		try
		{
		    SeldonDeployment existing = mlCache.get(mlDep.getMetadata().getName());
		    if (existing == null || !existing.getSpec().equals(mlDep.getSpec()))
		    {
		        logger.debug("Running updates for "+mlDep.getMetadata().getName());
		        SeldonDeployment mlDep2 = operator.defaulting(mlDep);
		        operator.validate(mlDep2);
		        mlCache.put(mlDep2);
		        DeploymentResources resources = operator.createResources(mlDep2);
		        ProtoClient client = clientProvider.getProtoClient();
		        createDeployments(client, resources.deployments);
		        removeDeployments(client,mlDep2,resources.deployments);
		        createService(client,resources.service);
		        if (!mlDep.getSpec().equals(mlDep2.getSpec()))
		        {
		            logger.debug("Pushing updated SeldonDeployment "+mlDep2.getMetadata().getName()+" back to kubectl");
		            crdHandler.updateSeldonDeployment(mlDep2);
		        }
		        else
		            logger.debug("Not pushing an update as no change to spec for SeldonDeployment "+mlDep2.getMetadata().getName());
		    }
		    else
		    {
		        mlCache.put(mlDep);
		        logger.debug("Only updated cache for "+mlDep.getMetadata().getName());
		    }
			
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
