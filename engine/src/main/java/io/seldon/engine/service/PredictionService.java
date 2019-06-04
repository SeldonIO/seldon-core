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

import java.io.IOException;
import java.math.BigInteger;
import java.security.SecureRandom;
import java.util.concurrent.ExecutionException;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.common.primitives.Doubles;
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

	@Value("${log.transform.tabular}")
	private boolean logTransformTabular;

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
			logMessageAsJson(feedback);
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
			//log json now we've added puid
			logMessageAsJson(request);
		}
		if(logResponses){
			logMessageAsJson(response);
		}

        return response;
		
	}

	private JsonNode transformJsonTabular(String json) throws IOException {
		ObjectMapper mapper = new ObjectMapper();
		JsonNode j = mapper.readTree(json);
		//only transform if there's a data element and it contains a names array
		if(j.has("data")&&j.get("data").has("names")) {
			JsonNode namesNode = j.get("data").get("names");

			String[] names = mapper.readValue(namesNode.toString(), String[].class);
			double[][] values = mapper.readValue(j.get("data").get("ndarray").toString(), double[][].class);
			double[] vs = Doubles.concat(values);

			for (int i = 0; i < names.length; i++) {
				((ObjectNode) j.get("data")).put(names[i], vs[i]);
			}

		}
		return j;
	}

	private void logMessageAsJson(SeldonMessage message){
		try {
			String json = ProtoBufUtils.toJson(message);
			if(logTransformTabular){
				json = transformJsonTabular(json).toString();
			}
			System.out.println(json);
		}catch (Exception ex){
			logger.error("Unable to parse message",ex);
		}
	}

	private void logMessageAsJson(Feedback message){
		try {
			String json = ProtoBufUtils.toJson(message);
			if(logTransformTabular){
				json = transformJsonTabular(json).toString();
			}
			System.out.println(json);
		}catch (Exception ex){
			logger.error("Unable to parse message",ex);
		}
	}
}
