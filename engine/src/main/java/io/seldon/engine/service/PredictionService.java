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
import java.util.ArrayList;
import java.util.concurrent.ExecutionException;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
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
		if(j.has("data")&&j.get("data").has("names") &&j.get("data").get("names").isArray()) {
			JsonNode namesNode = j.get("data").get("names");

			String[] names = mapper.readValue(namesNode.toString(), String[].class);
			//only transform if there's a data element and it contains a names array
			if(j.get("data").has("ndarray") && j.get("data").get("ndarray").isArray()) {

				ArrayList values = mapper.readValue(j.get("data").get("ndarray").toString(), ArrayList.class);

				//for 1D ndarray we map columns to features
				//we'll assume a 2D array is a batch where first dim is batches map columns per batch
				//for 3D we do nothing

				int nrDims =1;

				if(values.get(0).getClass().equals(ArrayList.class)){
					//array contains an array
					nrDims =2;
					ArrayList inner = (ArrayList)values.get(0);
					if(inner.get(0).getClass().equals(ArrayList.class)){
						nrDims=3;
						//too big - won't try to log

					}
				}

				if(nrDims==1){

					for (int i = 0; i < names.length; i++) {
						((ObjectNode) j.get("data")).set(names[i], mapper.readTree(mapper.writeValueAsString(values.get(i))));
					}

				} else if(nrDims==2){


					for (int i = 0; i < names.length; i++) {

						if(values.size()==1){

							//here we have a 2D array but one dimension is empty - so it's really single-dimension
							//store values as k-v rather than list
							((ObjectNode) j.get("data")).set(names[i], mapper.readTree(mapper.writeValueAsString(((ArrayList) values.get(0)).get(i))));

						} else {


							//really is a 2D array
							for (int row = 0; row < values.size(); row++) {
								ArrayList batchRow = (ArrayList) values.get(row);
								ArrayNode featureVals = null;
								if (j.get("data").has(names[i])) {
									featureVals = (ArrayNode) j.get("data").get(names[i]);
								} else {
									featureVals = mapper.createArrayNode();
								}
								featureVals.add(mapper.readTree(mapper.writeValueAsString(batchRow.get(i))));
								((ObjectNode) j.get("data")).set(names[i], featureVals);
							}
						}

					}
				}


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
