package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

import javax.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.util.JsonFormat;

import io.seldon.engine.zk.DeploymentsListener;
import io.seldon.engine.zk.ZkDeploymentsHandler;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

@Component
public class PredictorsStore implements DeploymentsListener {
	protected static Logger logger = LoggerFactory.getLogger(PredictorsStore.class.getName());
	public static final String ALG_KEY = "predict_algs";
	
	private ConcurrentMap<String, PredictorState> predictorsStore = new ConcurrentHashMap<>();
	
	@Autowired
	PredictorBean predictorBean;
	 
	 private final ZkDeploymentsHandler deploymentsHandler;
	 
	 @Autowired
	 public PredictorsStore(ZkDeploymentsHandler deploymentsHandler)
	 {	
		 this.deploymentsHandler = deploymentsHandler;
	 }
	 
	 @PostConstruct
	 private void init() throws Exception{
		 logger.info("Initializing...");
		 deploymentsHandler.addListener(this);
		 deploymentsHandler.contextInitialised();
	    }
	 
	 public PredictorState retrievePredictorState(String deployment)
	 {
		 PredictorState predictor = predictorsStore.get(deployment);
		 return predictor;
	 }
	 
	 
	 @Override
	 public void deploymentAdded(String deployment,String configValue) {
        ObjectMapper mapper = new ObjectMapper();
		logger.info("Detected new deployment: "+ deployment+": "+ configValue);
		try {
			
			//TODO: retrieve deploymentDef
			
			DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
	        DeploymentDef deploymentDef = null;
	        
			JsonFormat.parser().ignoringUnknownFields().merge(configValue, deploymentDefBuilder);
			deploymentDef = deploymentDefBuilder.build();

			PredictorState predictor = predictorBean.predictorStateFromDeploymentDef(deploymentDef);

			predictorsStore.put(deployment, predictor);
            logger.info("Succesfully updated predictor for "+ deployment);

        } catch (IOException | BeansException e) {
            logger.error("Couldn't update algorithms for deployment " +deployment, e);
        }
		
	}
	 
	@Override
	public void deploymentUpdated(String deployment,String configValue) {
		// SHOULD THIS REALLY HAPPEN AT EVERY CALL? WHY NOT A GLOBAL mapper? 
        ObjectMapper mapper = new ObjectMapper();
		logger.info("Received new algorithm config for "+ deployment+": "+ configValue);
		try {
			PredictorState predictor = mapper.readValue(configValue, PredictorState.class);
			predictorsStore.replace(deployment, predictor);
            logger.info("Succesfully updated predictor for "+ deployment);

        } catch (IOException | BeansException e) {
            logger.error("Couldn't update algorithms for deployment " +deployment, e);
        }
		
	}
	
	@Override
	public void deploymentRemoved(String deployment) {
		predictorsStore.remove(deployment);
		logger.info("Removed deployment "+deployment);
	}
		
}
