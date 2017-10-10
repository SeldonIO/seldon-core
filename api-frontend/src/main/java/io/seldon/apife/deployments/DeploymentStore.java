package io.seldon.apife.deployments;

import java.io.IOException;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

import javax.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.protobuf.util.JsonFormat;

import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

@Component
public class DeploymentStore implements DeploymentsListener {
	protected static Logger logger = LoggerFactory.getLogger(DeploymentStore.class.getName());
	
	//Oauth key to deployment def
	private ConcurrentMap<String, DeploymentDef> deploymentStore = new ConcurrentHashMap<>();
	
	private final DeploymentsHandler deploymentsHandler;
	 
	private InMemoryClientDetailsService ClientDetailsService;
	
	@Autowired
	public DeploymentStore(DeploymentsHandler deploymentsHandler,InMemoryClientDetailsService clientDetailsService)
	{	
		this.deploymentsHandler = deploymentsHandler;
		this.ClientDetailsService = clientDetailsService;
	}
	 
	@PostConstruct
	private void init() throws Exception{
		logger.info("Initializing...");
		deploymentsHandler.addListener(this);
	}
	 
	 public DeploymentDef getDeployment(String clientId)
	 {
		 return deploymentStore.get(clientId);
	 }
	 
	 
	 @Override
	 public void deploymentAdded(String resource) {
		logger.info("Detected new deployment: "+ resource);
		try {
			DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
	        DeploymentDef deploymentDef = null;
	        
			JsonFormat.parser().ignoringUnknownFields().merge(resource, deploymentDefBuilder);
			deploymentDef = deploymentDefBuilder.build();
			
			deploymentStore.put(deploymentDef.getOauthKey(), deploymentDef);
			ClientDetailsService.addClient(deploymentDef.getOauthKey(), deploymentDef.getOauthSecret());

			logger.info("Succesfully added deployment "+deploymentDef.getId());

        } catch (IOException | BeansException e) {
            logger.error("Couldn't update deployment " +resource, e);
        }
		
	}
	 
	@Override
	public void deploymentUpdated(String resource) {
		logger.info("Deployment updated "+resource);
		deploymentAdded(resource);
	}
	
	@Override
	public void deploymentRemoved(String resource) {
		try {
			DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
	        DeploymentDef deploymentDef = null;
	        
			JsonFormat.parser().ignoringUnknownFields().merge(resource, deploymentDefBuilder);
			deploymentDef = deploymentDefBuilder.build();

			deploymentStore.remove(deploymentDef.getOauthKey());
			logger.info("Removed deployment "+deploymentDef.getId());
		 } catch (IOException | BeansException e) {
	            logger.error("Couldn't delete deployment " +resource, e);
	        }
	}
		
}
