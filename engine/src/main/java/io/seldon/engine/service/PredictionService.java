/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.service;

import java.math.BigInteger;
import java.security.SecureRandom;
import java.util.concurrent.ExecutionException;

import io.seldon.engine.pb.ProtoBufUtils;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.predictors.PredictorBean;
import io.seldon.engine.predictors.PredictorState;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Meta;

@Service
public class PredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(PredictionService.class.getName());
	
	@Autowired
	PredictorBean predictorBean;
	
	@Autowired
	EnginePredictor enginePredictor;
	
	PuidGenerator puidGenerator = new PuidGenerator();

	@Value("${log.requests}")
	private boolean logRequests;

	@Value("${log.responses}")
	private boolean logResponses;

	@Value("${log.feedback.requests}")
	private boolean logFeedbackRequests;


	public final class PuidGenerator {
	    private SecureRandom random = new SecureRandom();

	    public String nextPuidId() {
	        return new BigInteger(130, random).toString(32);
	    }
	}
	
	public void sendFeedback(Feedback feedback) throws InterruptedException, ExecutionException, InvalidProtocolBufferException
	{
		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());

		predictorBean.sendFeedback(feedback, predictorState);

		if(logFeedbackRequests) {
			logRawMessageAsJson(feedback);
		}
		
		return;
	}
	
	public SeldonMessage predict(SeldonMessage request) throws InterruptedException, ExecutionException, InvalidProtocolBufferException
	{

		if (!request.hasMeta())
		{
			request = request.toBuilder().setMeta(Meta.newBuilder().setPuid(puidGenerator.nextPuidId()).build()).build();
		}
		else if (StringUtils.isEmpty(request.getMeta().getPuid()))
		{
			request = request.toBuilder().setMeta(request.getMeta().toBuilder().setPuid(puidGenerator.nextPuidId()).build()).build();
		}
		String puid = request.getMeta().getPuid();
		
        PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());

        SeldonMessage predictorReturn = predictorBean.predict(request,predictorState);
			
        SeldonMessage.Builder builder = SeldonMessage.newBuilder(predictorReturn).setMeta(Meta.newBuilder(predictorReturn.getMeta()).setPuid(puid));

        SeldonMessage response = builder.build();

		if(logRequests){
			//log pure json now we've added puid
			logRawMessageAsJson(request);
		}
		if(logResponses){
			logRawMessageAsJson(response);
		}

        return response;
		
	}

	private void logRawMessageAsJson(SeldonMessage message){
		try {
			System.out.println(ProtoBufUtils.toJson(message).replace("\n", "").replace("\r", ""));
		}catch (Exception ex){
			logger.error("Unable to parse message",ex);
		}
	}

	private void logRawMessageAsJson(Feedback message){
		try {
			System.out.println(ProtoBufUtils.toJson(message).replace("\n", "").replace("\r", ""));
		}catch (Exception ex){
			logger.error("Unable to parse message",ex);
		}
	}
}
