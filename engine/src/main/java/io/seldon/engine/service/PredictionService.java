package io.seldon.engine.service;

import java.math.BigInteger;
import java.security.SecureRandom;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.predictors.PredictorBean;
import io.seldon.engine.predictors.PredictorState;
import io.seldon.protos.PredictionProtos.FeedbackDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.MetaDef;

@Service
public class PredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(PredictionService.class.getName());
	
	private final ExecutorService pool = Executors.newFixedThreadPool(50);
	
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
	
	public void sendFeedback(FeedbackDef feedback) throws InterruptedException, ExecutionException
	{
		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());

		predictorBean.sendFeedback(feedback, predictorState);
		
		return;
	}
	
	public ResponseDef predict(RequestDef request) throws InterruptedException, ExecutionException
	{

		if (!request.hasMeta())
		{
			request = request.toBuilder().setMeta(MetaDef.newBuilder().setPuid(puidGenerator.nextPuidId()).build()).build();
		}
		else if (StringUtils.isEmpty(request.getMeta().getPuid()))
		{
			request = request.toBuilder().setMeta(request.getMeta().toBuilder().setPuid(puidGenerator.nextPuidId()).build()).build();
		}
		String puid = request.getMeta().getPuid();
		
        PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());

        ResponseDef predictorReturn = predictorBean.predict(request,predictorState);
			
        ResponseDef.Builder builder = ResponseDef.newBuilder(predictorReturn).setMeta(MetaDef.newBuilder(predictorReturn.getMeta()).setPuid(puid));

        return builder.build();
		
	}
}
