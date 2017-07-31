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
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
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
    public void testPredict_1dim() throws Exception
    {
        final String predictJson = "{" +
        	    "\"request\": {" + 
        	    "\"values\": [1.0,2.0]}" +
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
    public void testPredict_11dim() throws Exception
    {
        final String predictJson = "{" +
        	    "\"request\": {" + 
        	    "\"values\": [[1.0]]}" +
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
    public void testPredict_21dim() throws Exception
    {
        final String predictJson = "{" +
        	    "\"request\": {" + 
        	    "\"values\": [[1.0],[2.0]]}" +
        		"}";

    	MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
    }
}
