package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.apis.AutoscalingV2beta1Api;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1DeleteOptions;
import io.kubernetes.client.models.V1Service;
import io.kubernetes.client.models.V1ServiceList;
import io.kubernetes.client.models.V1Status;
import io.kubernetes.client.models.V2beta1HorizontalPodAutoscaler;
import io.kubernetes.client.models.V2beta1HorizontalPodAutoscalerList;
import io.kubernetes.client.proto.Meta.DeleteOptions;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.kubernetes.client.proto.V2beta1Autoscaling.HorizontalPodAutoscaler;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeletionHandler {

	private final static Logger logger = LoggerFactory.getLogger(SeldonDeletionHandler.class);
	
	private final SeldonNameCreator seldonNameCreator = new SeldonNameCreator();
	private final KubeCRDHandler crdHandler;
	
	
	@Autowired
	public SeldonDeletionHandler(KubeCRDHandler crdHandler) {
		super();
		this.crdHandler = crdHandler;
	}

	private Set<String> getServiceNames(List<Service> services)
	{
		Set<String> names = new HashSet<>();
		for(Service s : services)
			names.add(s.getMetadata().getName());
		return names;
	}
	
	private Set<String> getDeploymentNames(List<Deployment> deployments)
	{
		Set<String> names = new HashSet<>();
	    for(Deployment d : deployments)
	        names.add(d.getMetadata().getName());
	    return names;
	}
	
	private Set<String> getHpaNames(List<HorizontalPodAutoscaler> hpas)
	{
		Set<String> names = new HashSet<>();
	    for(HorizontalPodAutoscaler hpa : hpas)
	        names.add(hpa.getMetadata().getName());
	    return names;
	}
	
	
	/**
	 * Delete deployments that are not in list. Allows 2 stage delete by only deleting service orchestrator or all. Gets owned
	 * deployments and then removes ones not in the list provided.
	 * @param client  ProtoClient
	 * @param namespace Namespace to use
	 * @param seldonDeployment The Seldon Deployment we are refering to
	 * @param deployments The list of deployments from the Seldon Deployment
	 * @param svcOrchOnly Whether to only delete the service orchestrator
	 * @return Number of Deployments deleted
	 * @throws ApiException
	 * @throws IOException
	 * @throws SeldonDeploymentException
	 */
	public int removeDeployments(ProtoClient client,String namespace,SeldonDeployment seldonDeployment,List<Deployment> deployments,boolean svcOrchOnly) throws ApiException, IOException, SeldonDeploymentException
	{
		int deleteCount = 0;
	    Set<String> names = getDeploymentNames(deployments);
	    ExtensionsV1beta1DeploymentList depList = crdHandler.getOwnedDeployments(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
	    for (ExtensionsV1beta1Deployment d : depList.getItems())
	    {
	    	boolean okToDelete = !svcOrchOnly || (d.getMetadata().getLabels().containsKey(Constants.LABEL_SELDON_SVCORCH));
	        if (okToDelete && !names.contains(d.getMetadata().getName()))
	        {
	        	deleteCount++;
	            final String deleteApiPath = "/apis/"+SeldonDeploymentControllerImpl.DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
	                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(d.getMetadata().getName()))
	                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
	            DeleteOptions options = DeleteOptions.newBuilder().setPropagationPolicy("Foreground").build();
	            ObjectOrStatus<Deployment> os = client.delete(Deployment.newBuilder(),deleteApiPath,options);
	            if (os.status != null) {
                    logger.error("Error deleting deployment:"+ProtoBufUtils.toJson(os.status));
                    //throw new SeldonDeploymentException("Failed to delete deployment "+d.getMetadata().getName());
                }
                else {
                    logger.debug("Deleted deployment:"+ProtoBufUtils.toJson(os.object));
                }
	        }
	        else
	        	logger.info("Skipping deletion of {} svcOrchOnly:{}",d.getMetadata().getName(),svcOrchOnly);
	    }
	    return deleteCount;
	}
	
	public void removeServices(ApiClient client,String namespace,SeldonDeployment seldonDeployment,List<Service> services) throws ApiException, IOException, SeldonDeploymentException
	{
		Set<String> names = getServiceNames(services);
		V1ServiceList svcList = crdHandler.getOwnedServices(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
		for(V1Service s : svcList.getItems())
		{
			if (!names.contains(s.getMetadata().getName()))
			{	
				CoreV1Api api = new CoreV1Api(client);
				io.kubernetes.client.models.V1DeleteOptions options = new V1DeleteOptions();
				V1Status status = api.deleteNamespacedService(s.getMetadata().getName(), namespace, options, null, null, null, null);
				if (!"Success".equals(status.getStatus()))
				{
					logger.error("Failed to delete service "+s.getMetadata().getName());
					throw new SeldonDeploymentException("Failed to delete service "+s.getMetadata().getName());
				}
				else
					logger.debug("Deleted service "+s.getMetadata().getName());
				
			}
		}
	}
	
	public void removeHPAs(ApiClient client,String namespace,SeldonDeployment seldonDeployment,List<HorizontalPodAutoscaler> hpas) throws ApiException, IOException, SeldonDeploymentException
	{
		Set<String> names = getHpaNames(hpas);
		V2beta1HorizontalPodAutoscalerList hpaList = crdHandler.getOwnedHPAs(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
		for(V2beta1HorizontalPodAutoscaler hpa : hpaList.getItems())
		{
			if (!names.contains(hpa.getMetadata().getName()))
			{	
				AutoscalingV2beta1Api api = new AutoscalingV2beta1Api(client);
				io.kubernetes.client.models.V1DeleteOptions options = new V1DeleteOptions();
				V1Status status = api.deleteNamespacedHorizontalPodAutoscaler(hpa.getMetadata().getName(), namespace, options, null, null, null, null);
				if (!"Success".equals(status.getStatus()))
				{
					logger.error("Failed to delete HPA "+hpa.getMetadata().getName());
					throw new SeldonDeploymentException("Failed to delete HPA "+hpa.getMetadata().getName());
				}
				else
					logger.debug("Deleted HPA "+hpa.getMetadata().getName());
				
			}
		}
	}
}
