package io.seldon.engine.api.rest;

import java.net.URI;
import java.nio.charset.StandardCharsets;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Matchers;
import org.mockito.Mockito;
import org.mockito.invocation.InvocationOnMock;
import org.mockito.stubbing.Answer;
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
import org.springframework.util.MultiValueMap;
import org.springframework.web.context.WebApplicationContext;

import io.seldon.engine.filters.XSSFilter;
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
public class TestRestClientControllerExternalGraphs {

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
        .addFilters(new XSSFilter())
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
    	String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_with_metrics.json",StandardCharsets.UTF_8);
    	String responseStr2 = readFile("src/test/resources/response_with_metrics2.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"data\": {" +
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);


    	ResponseEntity<String> httpResponse1 = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	ResponseEntity<String> httpResponse2 = new ResponseEntity<String>(responseStr2, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenAnswer(new Answer<ResponseEntity<String>>() {
    		    private int count = 0;

    		    public ResponseEntity<String> answer(InvocationOnMock invocation) {
    		    	count++;
    		        if (count == 1)
    		            return httpResponse1;

    		        return httpResponse2;
    		    }});

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer",seldonMessage.getMeta().getMetrics(2).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    System.out.println("response is ["+response+"]");
	    Assert.assertTrue(response.indexOf("mycounter_total{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mygauge{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 22.0")>-1);
    	System.out.println(response);

    	res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());

    	builder = SeldonMessage.newBuilder();
	    JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
	    seldonMessage = builder.build();

    	 // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(100.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer",seldonMessage.getMeta().getMetrics(2).getKey());

	 // Check prometheus endpoint for metric
	    res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    Assert.assertTrue(response.indexOf("mycounter_total{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 2.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 2.0")>-1);
	    Assert.assertTrue(response.indexOf("mygauge{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 100.0")>-1);
    	System.out.println(response);
    }


    @Test
    public void testInputTransformInputMetrics() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/transformer_simple.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_with_metrics.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"data\": {" +
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);


    	ResponseEntity<String> httpResponse = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenReturn(httpResponse);

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer",seldonMessage.getMeta().getMetrics(2).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    Assert.assertTrue(response.indexOf("mycounter_total{deployment_name=\"None\",model_image=\"seldonio/transformer\",model_name=\"transformer\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/transformer\",model_name=\"transformer\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
    	System.out.println(response);
    }


    @Test
    public void testTransformOutputMetrics() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/transform_output_simple.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_with_metrics.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"data\": {" +
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);


    	ResponseEntity<String> httpResponse = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenReturn(httpResponse);

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer",seldonMessage.getMeta().getMetrics(2).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
    	System.out.println(response);
	    Assert.assertTrue(response.indexOf("mycounter_total{deployment_name=\"None\",model_image=\"seldonio/transformer\",model_name=\"transform_output\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/transformer\",model_name=\"transform_output\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);

    }


    @Test
    public void testRouterMetrics() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/router_simple.json",StandardCharsets.UTF_8);
    	String responseStrRouter = readFile("src/test/resources/router_response.json",StandardCharsets.UTF_8);
    	String responseStrModel = readFile("src/test/resources/router_model_response.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"data\": {" +
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);


    	ResponseEntity<String> httpResponse1 = new ResponseEntity<String>(responseStrRouter, null, HttpStatus.OK);
    	ResponseEntity<String> httpResponse2 = new ResponseEntity<String>(responseStrModel, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
		.thenAnswer(new Answer<ResponseEntity<String>>() {
		    private int count = 0;

		    public ResponseEntity<String> answer(InvocationOnMock invocation) {
		    	count++;
		        if (count == 1)
		            return httpResponse1;

		        return httpResponse2;
		    }});

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("myroutercounter",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("myroutergauge",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("myroutertimer",seldonMessage.getMeta().getMetrics(2).getKey());

	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(3).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(3).getValue(),0.0);
	    Assert.assertEquals("myroutermodelcounter",seldonMessage.getMeta().getMetrics(3).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(4).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(4).getValue(),0.0);
	    Assert.assertEquals("myroutermodelgauge",seldonMessage.getMeta().getMetrics(4).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(5).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(5).getValue(),0.0);
	    Assert.assertEquals("myroutermodeltimer",seldonMessage.getMeta().getMetrics(5).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
    	System.out.println(response);
	    Assert.assertTrue(response.indexOf("myroutercounter_total{deployment_name=\"None\",model_image=\"seldonio/router\",model_name=\"router\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("myroutertimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/router\",model_name=\"router\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("myroutermodelcounter_total{deployment_name=\"None\",model_image=\"seldonio/model\",model_name=\"model\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("myroutermodeltimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/model\",model_name=\"model\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);

    }


    @Test
    public void testCombinerMetrics() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/combiner_simple.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_with_metrics.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"data\": {" +
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);


    	ResponseEntity<String> httpResponse = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenReturn(httpResponse);

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer",seldonMessage.getMeta().getMetrics(2).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    System.out.println("----------------------------------------");
	    System.out.println("----------------------------------------");
	    System.out.println(response);
	    Assert.assertTrue(response.indexOf("mycounter_total{deployment_name=\"None\",model_image=\"seldonio/combiner\",model_name=\"combiner\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer_seconds_count{deployment_name=\"None\",model_image=\"seldonio/combiner\",model_name=\"combiner\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);

    }


    @Test
    public void testModelStrData() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_strdata.json",StandardCharsets.UTF_8);
    	String responseStr2 = readFile("src/test/resources/response_strdata2.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"strData\": \"my string data\"" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);

    	ResponseEntity<String> httpResponse1 = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	ResponseEntity<String> httpResponse2 = new ResponseEntity<String>(responseStr2, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenAnswer(new Answer<ResponseEntity<String>>() {
    		    private int count = 0;

    		    public ResponseEntity<String> answer(InvocationOnMock invocation) {
    		    	count++;
    		        if (count == 1)
    		            return httpResponse1;

    		        return httpResponse2;
    		    }});

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter1",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge1",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer1",seldonMessage.getMeta().getMetrics(2).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    System.out.println("response is ["+response+"]");
	    Assert.assertTrue(response.indexOf("mycounter1_total{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer1_seconds_count{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mygauge1{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 22.0")>-1);
    	System.out.println(response);

    	res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());

    	builder = SeldonMessage.newBuilder();
	    JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
	    seldonMessage = builder.build();

    	 // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter1",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(100.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge1",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer1",seldonMessage.getMeta().getMetrics(2).getKey());

	 // Check prometheus endpoint for metric
	    res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    Assert.assertTrue(response.indexOf("mycounter1_total{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 2.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer1_seconds_count{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 2.0")>-1);
	    Assert.assertTrue(response.indexOf("mygauge1{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 100.0")>-1);
    	System.out.println(response);
    }


    @Test
    public void testModelBinData() throws Exception
    {
    	String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
    	String responseStr = readFile("src/test/resources/response_bindata.json",StandardCharsets.UTF_8);
    	String responseStr2 = readFile("src/test/resources/response_bindata2.json",StandardCharsets.UTF_8);
    	PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
    	EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
    	PredictorSpec predictorSpec = PredictorSpecBuilder.build();
    	final String predictJson = "{" +
         	    "\"binData\": \"MTIz\"" +
         		"}";
    	ReflectionTestUtils.setField(enginePredictor,"predictorSpec",predictorSpec);

    	ResponseEntity<String> httpResponse1 = new ResponseEntity<String>(responseStr, null, HttpStatus.OK);
    	ResponseEntity<String> httpResponse2 = new ResponseEntity<String>(responseStr2, null, HttpStatus.OK);
    	Mockito.when(testRestTemplate.getRestTemplate().postForEntity(Matchers.<URI>any(), Matchers.<HttpEntity<MultiValueMap<String, String>>>any(), Matchers.<Class<String>>any()))
    		.thenAnswer(new Answer<ResponseEntity<String>>() {
    		    private int count = 0;

    		    public ResponseEntity<String> answer(InvocationOnMock invocation) {
    		    	count++;
    		        if (count == 1)
    		            return httpResponse1;

    		        return httpResponse2;
    		    }});

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

	    // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter2",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(22.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge2",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer2",seldonMessage.getMeta().getMetrics(2).getKey());

	    // Check prometheus endpoint for metric
	    MvcResult res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    System.out.println("response is ["+response+"]");
	    Assert.assertTrue(response.indexOf("mycounter2_total{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer2_seconds_count{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 1.0")>-1);
	    Assert.assertTrue(response.indexOf("mygauge2{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 22.0")>-1);
    	System.out.println(response);

    	res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());

    	builder = SeldonMessage.newBuilder();
	    JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
	    seldonMessage = builder.build();

    	 // Check for returned metrics
	    Assert.assertEquals("COUNTER",seldonMessage.getMeta().getMetrics(0).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(0).getValue(),0.0);
	    Assert.assertEquals("mycounter2",seldonMessage.getMeta().getMetrics(0).getKey());

	    Assert.assertEquals("GAUGE",seldonMessage.getMeta().getMetrics(1).getType().toString());
	    Assert.assertEquals(100.0F,seldonMessage.getMeta().getMetrics(1).getValue(),0.0);
	    Assert.assertEquals("mygauge2",seldonMessage.getMeta().getMetrics(1).getKey());

	    Assert.assertEquals("TIMER",seldonMessage.getMeta().getMetrics(2).getType().toString());
	    Assert.assertEquals(1.0F,seldonMessage.getMeta().getMetrics(2).getValue(),0.0);
	    Assert.assertEquals("mytimer2",seldonMessage.getMeta().getMetrics(2).getKey());

	 // Check prometheus endpoint for metric
	    res2 = mvc.perform(MockMvcRequestBuilders.get("/prometheus")).andReturn();
	    Assert.assertEquals(200, res2.getResponse().getStatus());
	    response = res2.getResponse().getContentAsString();
	    Assert.assertTrue(response.indexOf("mycounter2_total{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",mytag1=\"mytagval1\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 2.0")>-1);
	    Assert.assertTrue(response.indexOf("mytimer2_seconds_count{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 2.0")>-1);
	    Assert.assertTrue(response.indexOf("mygauge2{deployment_name=\"None\",model_image=\"seldonio/mock_classifier\",model_name=\"mean-classifier\",model_version=\"0.6\",predictor_name=\"fx-market-predictor\",predictor_version=\"unknown\",} 100.0")>-1);
    	System.out.println(response);
    }

}
