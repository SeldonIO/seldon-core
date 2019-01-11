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
package io.seldon.engine.api.rest;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.embedded.LocalServerPort;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.http.MediaType;
import org.springframework.jmx.support.MetricType;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
//@AutoConfigureMockMvc
public class TestRestClientController {
	
	@Autowired
	private WebApplicationContext context;
	
	
    //@Autowired
    private MockMvc mvc;
    
    
    @Before
	public void setup() {
		mvc = MockMvcBuilders
				.webAppContextSetup(context)
				.build();
	}
    
    @LocalServerPort
    private int port;


    @Test
    public void testPing() throws Exception
    {
    	MvcResult res = mvc.perform(MockMvcRequestBuilders.get("/ping")).andReturn();
    	String response = res.getResponse().getContentAsString();
    	Assert.assertEquals("pong", response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
    }
    
    
    
    @Test
    public void testPredict_11dim_ndarry() throws Exception
    {
        final String predictJson = "{" +
        	    "\"request\": {" + 
        	    "\"ndarray\": [[1.0]]}" +
        		"}";

    	MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
    }
    	    
    @Test
    public void testPredict_21dim_ndarry() throws Exception
    {
        final String predictJson = "{" +
        	    "\"request\": {" + 
        	    "\"ndarray\": [[1.0],[2.0]]}" +
        		"}";

    	MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
    	SeldonMessage.Builder builder = SeldonMessage.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, response );
		SeldonMessage seldonMessage = builder.build();
		Assert.assertEquals(3, seldonMessage.getMeta().getMetricsCount());
		Assert.assertEquals("COUNTER", seldonMessage.getMeta().getMetrics(0).getType().toString());
		Assert.assertEquals("GAUGE", seldonMessage.getMeta().getMetrics(1).getType().toString());
		Assert.assertEquals("TIMER", seldonMessage.getMeta().getMetrics(2).getType().toString());
    }
    
    @Test
    public void testPredict_21dim_tensor() throws Exception
    {
        final String predictJson = "{" +
        	    "\"request\": {" + 
        	    "\"tensor\": {\"shape\":[2,1],\"values\":[1.0,2.0]}}" +
        		"}";

    	MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
    	SeldonMessage.Builder builder = SeldonMessage.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, response );
		SeldonMessage seldonMessage = builder.build();
		Assert.assertEquals(3, seldonMessage.getMeta().getMetricsCount());
		Assert.assertEquals("COUNTER", seldonMessage.getMeta().getMetrics(0).getType().toString());
		Assert.assertEquals("GAUGE", seldonMessage.getMeta().getMetrics(1).getType().toString());
		Assert.assertEquals("TIMER", seldonMessage.getMeta().getMetrics(2).getType().toString());
    }
}
