package io.seldon.engine.api.rest;

import java.net.URI;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Matchers;
import org.mockito.Mockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.web.server.LocalServerPort;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.TestPropertySource;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers;
import org.springframework.util.MultiValueMap;
import org.springframework.web.context.WebApplicationContext;

import io.seldon.engine.pb.JsonFormat;
import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.SeldonMessage;

import static io.seldon.engine.util.TestUtils.readFile;

@RunWith(SpringRunner.class)
@ActiveProfiles("test")
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
@TestPropertySource(properties = {
	    "management.security.enabled=false",
	})
public class TestRandomABTest {

	@Autowired
	private WebApplicationContext context;

	@Autowired
	private EnginePredictor enginePredictor;


    //@Autowired
    private MockMvc mvc;

    @Autowired
    private RestClientController restController;

    @Before
	public void setup() throws Exception {
    	mvc = MockMvcBuilders
				.webAppContextSetup(context)
        .apply(SecurityMockMvcConfigurers.springSecurity())
				.build();
	}

    @LocalServerPort
    private int port;

    @Autowired
    private TestRestTemplate testRestTemplate;

    @Autowired
    private InternalPredictionService internalPredictionService;



    @Test
    public void testModelMetrics() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/abtest.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_with_metrics.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"request\": {" +
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);

    	ResponseEntity<String> httpResponse = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenReturn(httpResponse);

    	int routeACount = 0;
    	int routeBCount = 1;

    	for(int i=0;i<100;i++)
    	{
    		MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
        			.accept(MediaType.APPLICATION_JSON_UTF8)
        			.content(predictJson)
        			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
        	String response = res.getResponse().getContentAsString();
        	System.out.println(response);
        	Assert.assertEquals(200, res.getResponse().getStatus());

        	SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    	    JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
    	    SeldonMessage seldonMessage = builder.build();

    	    Assert.assertTrue(seldonMessage.getMeta().getRoutingMap().get("abtest") >= 0);
    	    if (seldonMessage.getMeta().getRoutingMap().get("abtest") == 0)
    	    	routeACount++;
    	    else
    	    	routeBCount++;
    	}
    	double split = routeACount /(double)(routeACount + routeBCount);
 	    System.out.println(routeACount);
 	    System.out.println(routeBCount);
 	    Assert.assertEquals(0.5, split,0.2);



    }

}
