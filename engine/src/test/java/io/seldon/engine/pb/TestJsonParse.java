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
package io.seldon.engine.pb;

import java.io.IOException;
import java.util.ArrayList;

import com.fasterxml.jackson.databind.node.ArrayNode;
import org.junit.Assert;
import org.junit.Test;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.common.primitives.Doubles;

public class TestJsonParse {

	String rawRequestSingleDim = "{  \"meta\": {    \"puid\": \"avodt6jrk9nbgomnco7nhrvpo0\",    \"tags\": {    },    \"routing\": {    },    \"requestPath\": {    },    \"metrics\": []  },  \"data\": {    \"names\": [\"f0\", \"f1\"],    \"ndarray\": [0.15, 0.74]  }}";
	String rawRequest = "{  \"meta\": {    \"puid\": \"avodt6jrk9nbgomnco7nhrvpo0\",    \"tags\": {    },    \"routing\": {    },    \"requestPath\": {    },    \"metrics\": []  },  \"data\": {    \"names\": [\"f0\", \"f1\"],    \"ndarray\": [[0.15, 0.74]]  }}";
	String rawRequestBatch = "{  \"meta\": {    \"puid\": \"avodt6jrk9nbgomnco7nhrvpo0\",    \"tags\": {    },    \"routing\": {    },    \"requestPath\": {    },    \"metrics\": []  },  \"data\": {    \"names\": [\"f0\", \"f1\"],    \"ndarray\": [[0.15, 0.74],[0.16, 0.75]]  }}";
	String rawResponse = "{  \"meta\": {    \"puid\": \"avodt6jrk9nbgomnco7nhrvpo0\",    \"tags\": {    },    \"routing\": {    },    \"requestPath\": {      \"classifier\": \"seldonio/mock_classifier:1.0\"    },    \"metrics\": []  },  \"data\": {    \"names\": [\"proba\"],    \"ndarray\": [[0.07786847593954888]]  }}";
	
	@Test
	public void multiDimTest() throws JsonProcessingException, IOException
	{
		String json = "{\"request\":{\"values\":[[1.0]]}}";

		System.out.println(json);
		ObjectMapper mapper = new ObjectMapper();
		JsonFactory factory = mapper.getFactory();
		JsonParser parser = factory.createParser(json);
		JsonNode j = mapper.readTree(parser);
		JsonNode values = j.get("request").get("values");
		
		double[][] v = mapper.readValue(values.toString(),double[][].class);
		double[] vs = Doubles.concat(v);
		int[] shape = {v.length,v[0].length };
		((ObjectNode) j.get("request")).replace("values",mapper.valueToTree(vs));
		((ObjectNode) j.get("request")).set("shape",mapper.valueToTree(shape));
		System.out.println(j.toString());
	}

	@Test
	public void tabularTransformRequestSingleDimTest() throws JsonProcessingException, IOException
	{

		JsonNode j = transformJson(rawRequestSingleDim);
		//f0 is 0.15
		Assert.assertEquals(j.toString(),"{\"meta\":{\"puid\":\"avodt6jrk9nbgomnco7nhrvpo0\",\"tags\":{},\"routing\":{},\"requestPath\":{},\"metrics\":[]},\"data\":{\"names\":[\"f0\",\"f1\"],\"ndarray\":[0.15,0.74],\"f0\":0.15,\"f1\":0.74}}");
	}
	@Test
	public void tabularTransformRequestTest() throws JsonProcessingException, IOException
	{

		JsonNode j = transformJson(rawRequest);
		//f0 is 0.15 - it is batch form of 2D but only contains one entry (a batch of one)
		// want f0 to get straight value rather than array with one entry
		Assert.assertEquals(j.toString(),"{\"meta\":{\"puid\":\"avodt6jrk9nbgomnco7nhrvpo0\",\"tags\":{},\"routing\":{},\"requestPath\":{},\"metrics\":[]},\"data\":{\"names\":[\"f0\",\"f1\"],\"ndarray\":[[0.15,0.74]],\"f0\":0.15,\"f1\":0.74}}");
	}

	@Test
	public void tabularTransformBatchRequestTest() throws JsonProcessingException, IOException
	{

		JsonNode j = transformJson(rawRequestBatch);
		//f0 should be array with 0.15 from first entry and 0.16 from second
		Assert.assertEquals(j.toString(),"{\"meta\":{\"puid\":\"avodt6jrk9nbgomnco7nhrvpo0\",\"tags\":{},\"routing\":{},\"requestPath\":{},\"metrics\":[]},\"data\":{\"names\":[\"f0\",\"f1\"],\"ndarray\":[[0.15,0.74],[0.16,0.75]],\"f0\":[0.15,0.16],\"f1\":[0.74,0.75]}}");
	}

	@Test
	public void tabularTransformResponseTest() throws JsonProcessingException, IOException
	{

		JsonNode j = transformJson(rawResponse);
		//only entry is proba and it has one value
		Assert.assertEquals(j.toString(),"{\"meta\":{\"puid\":\"avodt6jrk9nbgomnco7nhrvpo0\",\"tags\":{},\"routing\":{},\"requestPath\":{\"classifier\":\"seldonio/mock_classifier:1.0\"},\"metrics\":[]},\"data\":{\"names\":[\"proba\"],\"ndarray\":[[0.07786847593954888]],\"proba\":0.07786847593954888}}");
	}

	private JsonNode transformJson(String json) throws IOException {
		ObjectMapper mapper = new ObjectMapper();
		JsonNode j = mapper.readTree(json);
		//only transform if there's a data element and it contains a names array
		if(j.has("data")&&j.get("data").has("names") &&j.get("data").get("names").isArray()) {
			JsonNode namesNode = j.get("data").get("names");

			String[] names = mapper.readValue(namesNode.toString(), String[].class);
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
}
