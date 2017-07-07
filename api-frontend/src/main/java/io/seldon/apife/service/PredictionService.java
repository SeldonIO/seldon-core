package io.seldon.apife.service;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;

import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.apife.exception.APIException;
import io.seldon.protos.DeploymentProtos.DeploymentDef;


@Service
public class PredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(PredictionService.class.getName());
		
	@Autowired
	DeploymentStore deploymentStore;
	
	@Autowired
	InternalPredictionService internalPredictionService;
	

	public JsonNode predict(JsonNode request,String clientId)  {

		DeploymentDef deployment = deploymentStore.getDeployment(clientId);
		if (deployment != null)
			return internalPredictionService.getPrediction(request, deployment.getPredictor().getEndpoint());
		else
			throw new APIException(APIException.NO_RUNNING_DEPLOYMENT);

	}
}
