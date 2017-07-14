package io.seldon.engine.service;

import java.io.IOException;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.seldon.engine.exception.APIException;
import io.seldon.engine.logging.PredictLogger;
import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.predictors.PredictorBean;
import io.seldon.engine.predictors.PredictorRequest;
import io.seldon.engine.predictors.PredictorReturn;
import io.seldon.engine.predictors.PredictorState;
import java.security.SecureRandom;
import java.math.BigInteger;

@Service
public class PredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(PredictionService.class.getName());
	
	private final ExecutorService pool = Executors.newFixedThreadPool(50);
	
	@Autowired
	PredictLogger predictLogger;
	
//	@Autowired
//	PredictorsStore predictorsStore;
	
	@Autowired
	PredictorBean predictorBean;
	
	@Autowired
	EnginePredictor enginePredictor;
	
	PuidGenerator puidGenerator = new PuidGenerator();

	public final class PuidGenerator {
	    private SecureRandom random = new SecureRandom();

	    public String nextPuidId() {
	        return new BigInteger(130, random).toString(32);
	    }
	}
	
	private JsonNode getValidatedJson(String jsonRaw) throws JsonParseException, IOException
	{
		ObjectMapper mapper = new ObjectMapper();
	    JsonFactory factory = mapper.getFactory();
	    JsonParser parser = factory.createParser(jsonRaw);
	    JsonNode actualObj = mapper.readTree(parser);
	    
	    return actualObj;
	}
	
	public PredictionServiceReturn predict(PredictionServiceRequest predictionServiceRequest) throws APIException, InterruptedException, ExecutionException{

		if (StringUtils.isEmpty(predictionServiceRequest.getMeta().getPuid()))
			predictionServiceRequest.getMeta().setPuid(puidGenerator.nextPuidId());
		
        PredictorState predictorState = predictorBean.predictorStateFromDeploymentDef(enginePredictor.getPredictorDef());

        PredictorReturn predictorReturn = predictorBean.predict(predictionServiceRequest,predictorState);
			
        PredictionServiceReturnMeta meta = new PredictionServiceReturnMeta(predictionServiceRequest.getMeta().getPuid());
			
        PredictionServiceReturn res = new PredictionServiceReturn(meta, predictorReturn);
			
        //predictLogger.log(predictionServiceRequest.meta.deployment, predictionServiceRequest.request, res);
        return res;
		
	}
}
