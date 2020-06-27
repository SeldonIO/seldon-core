/**
 * ***************************************************************************** Copyright 2017
 * Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * <p>Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 *
 * <p>http://www.apache.org/licenses/LICENSE-2.0
 *
 * <p>Unless required by applicable law or agreed to in writing, software distributed under the
 * License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 * *****************************************************************************
 */
package io.seldon.engine.pb;

import static java.time.format.DateTimeFormatter.ISO_ZONED_DATE_TIME;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.common.primitives.Doubles;
import java.io.IOException;
import java.time.ZonedDateTime;
import org.junit.Assert;
import org.junit.Test;

public class TestJsonParse {

  String rawRequest =
      "{  \"meta\": {    \"puid\": \"avodt6jrk9nbgomnco7nhrvpo0\",    \"tags\": {    },    \"routing\": {    },    \"requestPath\": {    },    \"metrics\": []  },  \"data\": {    \"names\": [\"f0\", \"f1\"],    \"ndarray\": [[0.15, 0.74]]  }}";
  String rawResponse =
      "{  \"meta\": {    \"puid\": \"avodt6jrk9nbgomnco7nhrvpo0\",    \"tags\": {    },    \"routing\": {    },    \"requestPath\": {      \"classifier\": \"seldonio/mock_classifier:1.0\"    },    \"metrics\": []  },  \"data\": {    \"names\": [\"proba\"],    \"ndarray\": [[0.07786847593954888]]  }}";

  @Test
  public void multiDimTest() throws JsonProcessingException, IOException {
    String json = "{\"request\":{\"values\":[[1.0]]}}";

    ObjectMapper mapper = new ObjectMapper();
    JsonFactory factory = mapper.getFactory();
    JsonParser parser = factory.createParser(json);
    JsonNode j = mapper.readTree(parser);
    JsonNode values = j.get("request").get("values");

    double[][] v = mapper.readValue(values.toString(), double[][].class);
    double[] vs = Doubles.concat(v);
    int[] shape = {v.length, v[0].length};
    ((ObjectNode) j.get("request")).replace("values", mapper.valueToTree(vs));
    ((ObjectNode) j.get("request")).set("shape", mapper.valueToTree(shape));
  }

  private JsonNode combineRequestResponse(
      String request, String response, ZonedDateTime requestTime, ZonedDateTime responseTime)
      throws IOException {
    ObjectMapper mapper = new ObjectMapper();
    JsonNode requestNode = mapper.readTree(request);
    JsonNode responseNode = mapper.readTree(response);
    ObjectNode combined = mapper.createObjectNode();
    combined.set("request", requestNode);
    combined.set("response", responseNode);
    ((ObjectNode) combined.get("request"))
        .set("date", mapper.readTree(mapper.writeValueAsString(requestTime.toString())));
    ((ObjectNode) combined.get("response"))
        .set("date", mapper.readTree(mapper.writeValueAsString(responseTime.toString())));
    String depName = System.getenv().get("DEPLOYMENT_NAME");
    if (depName != null) {
      combined.set("sdepName", mapper.readTree(depName));
    }

    return combined;
  }

  @Test
  public void combineRequestResponse() throws JsonProcessingException, IOException {

    ZonedDateTime time = ZonedDateTime.parse("2018-04-26T14:48:09.769Z", ISO_ZONED_DATE_TIME);
    JsonNode j = combineRequestResponse(rawRequest, rawResponse, time, time);
    Assert.assertEquals(
        j.toString(),
        "{\"request\":{\"meta\":{\"puid\":\"avodt6jrk9nbgomnco7nhrvpo0\",\"tags\":{},\"routing\":{},\"requestPath\":{},\"metrics\":[]},\"data\":{\"names\":[\"f0\",\"f1\"],\"ndarray\":[[0.15,0.74]]},\"date\":\"2018-04-26T14:48:09.769Z\"},\"response\":{\"meta\":{\"puid\":\"avodt6jrk9nbgomnco7nhrvpo0\",\"tags\":{},\"routing\":{},\"requestPath\":{\"classifier\":\"seldonio/mock_classifier:1.0\"},\"metrics\":[]},\"data\":{\"names\":[\"proba\"],\"ndarray\":[[0.07786847593954888]]},\"date\":\"2018-04-26T14:48:09.769Z\"}}");
  }
}
