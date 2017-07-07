package io.seldon.apife.deployments;

import java.io.IOException;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

import javax.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.oauth2.provider.ClientDetailsService;
import org.springframework.stereotype.Component;

import com.google.protobuf.util.JsonFormat;

import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.apife.zk.DeploymentsListener;
import io.seldon.apife.zk.ZkDeploymentsHandler;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

@Component
public class DeploymentStore implements DeploymentsListener {
	protected static Logger logger = LoggerFactory.getLogger(DeploymentStore.class.getName());
	
	//Oauth key to deployment def
	private ConcurrentMap<String, DeploymentDef> deploymentStore = new ConcurrentHashMap<>();
	
	private final ZkDeploymentsHandler deploymentsHandler;
	 
	private InMemoryClientDetailsService ClientDetailsService;
	
	@Autowired
	public DeploymentStore(ZkDeploymentsHandler deploymentsHandler,InMemoryClientDetailsService clientDetailsService)
	{	
		this.deploymentsHandler = deploymentsHandler;
		this.ClientDetailsService = clientDetailsService;
	}
	 
	@PostConstruct
	private void init() throws Exception{
		logger.info("Initializing...");
		deploymentsHandler.addListener(this);
		deploymentsHandler.contextInitialised();
	}
	 
	 public DeploymentDef getDeployment(String clientId)
	 {
		 return deploymentStore.get(clientId);
	 }
	 
	 
	 @Override
	 public void deploymentAdded(String deployment,String configValue) {
		logger.info("Detected new deployment: "+ deployment+": "+ configValue);
		try {
			DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
	        DeploymentDef deploymentDef = null;
	        
			JsonFormat.parser().ignoringUnknownFields().merge(configValue, deploymentDefBuilder);
			deploymentDef = deploymentDefBuilder.build();
			
			deploymentStore.put(deploymentDef.getOauthKey(), deploymentDef);
			ClientDetailsService.addClient(deploymentDef.getOauthKey(), deploymentDef.getOauthSecret());

			logger.info("Succesfully updated predictor for "+ deployment);

        } catch (IOException | BeansException e) {
            logger.error("Couldn't update deployment " +deployment, e);
        }
		
	}
	 
	@Override
	public void deploymentUpdated(String deployment,String configValue) {
		logger.info("Deployment updated "+deployment);
		deploymentAdded(deployment, configValue);
	}
	
	@Override
	public void deploymentRemoved(String deployment) {
		deploymentStore.remove(deployment);
		logger.info("Removed deployment "+deployment);
	}
		
}
