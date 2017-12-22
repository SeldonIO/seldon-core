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

import org.junit.Test;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.common.primitives.Doubles;

public class TestJsonParse {

	
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
}
