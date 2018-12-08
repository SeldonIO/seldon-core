package io.seldon.apife.rest;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.security.Principal;

import org.apache.http.Header;
import org.apache.http.HttpEntity;
import org.apache.http.ProtocolVersion;
import org.apache.http.StatusLine;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.protocol.HttpContext;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.embedded.LocalServerPort;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.http.MediaType;
import org.springframework.test.context.TestPropertySource;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.context.WebApplicationContext;

import io.seldon.apife.SeldonTestBase;
import io.seldon.apife.api.rest.RestClientController;
import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.apife.service.InternalPredictionService;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
@TestPropertySource(properties = {
	    "management.security.enabled=false",
	})
public class RestControllerTest extends SeldonTestBase {
	
	public static class MockStatusLine implements StatusLine
	{
		private final int respCode;
		
		public MockStatusLine(int respCode) {
			this.respCode = respCode;
		}

		@Override
		public ProtocolVersion getProtocolVersion() {
			// TODO Auto-generated method stub
			return null;
		}

		@Override
		public int getStatusCode() {
			return this.respCode;
		}

		@Override
		public String getReasonPhrase() {
			// TODO Auto-generated method stub
			return null;
		}
		
	}
	
	public static class MockHttpEntity implements HttpEntity
	{
		private final String respStr;

		public MockHttpEntity(String respStr) {
			super();
			this.respStr = respStr;
		}

		@Override
		public boolean isRepeatable() {
			// TODO Auto-generated method stub
			return false;
		}

		@Override
		public boolean isChunked() {
			// TODO Auto-generated method stub
			return false;
		}

		@Override
		public long getContentLength() {
			// TODO Auto-generated method stub
			return 0;
		}

		@Override
		public Header getContentType() {
			// TODO Auto-generated method stub
			return null;
		}

		@Override
		public Header getContentEncoding() {
			// TODO Auto-generated method stub
			return null;
		}

		@Override
		public InputStream getContent() throws IOException, UnsupportedOperationException {
			return new ByteArrayInputStream(respStr.getBytes());
		}

		@Override
		public void writeTo(OutputStream outstream) throws IOException {
			// TODO Auto-generated method stub
			
		}

		@Override
		public boolean isStreaming() {
			// TODO Auto-generated method stub
			return false;
		}

		@Override
		public void consumeContent() throws IOException {
			// TODO Auto-generated method stub
			
		}
		
	}
	
	@Autowired
	private WebApplicationContext context;
	
		
	//@Autowired
	private MockMvc mvc;
  
	@Autowired
	RestClientController restController;
	
	@Before
		public void setup() throws Exception {
  
		mvc = MockMvcBuilders
				.webAppContextSetup(context)
				.build();
	}
  
	@LocalServerPort
	private int port;
  
	@Mock
	private RestTemplate restTemplate;

	@Autowired
	private InternalPredictionService internalPredictionService;

	@Autowired
	private DeploymentStore deploymentStore;
	
	@Test
	public void invalidClientTest() throws Exception
	{
		final String predictJson = "{" +
         	    "\"data\": {" + 
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
		
		Principal mockPrincipal = Mockito.mock(Principal.class);
	    Mockito.when(mockPrincipal.getName()).thenReturn("invalidClientId");
		
		MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
				.principal(mockPrincipal)
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(500, res.getResponse().getStatus());
	}
	
	@Test
	public void predictBadResponseTest() throws Exception
	{
		final String predictJson = "{" +
         	    "\"data\": {" + 
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
		SeldonDeployment.Builder dBuilder = SeldonDeployment.newBuilder();
		updateMessageBuilderFromJson(dBuilder, jsonStr);
		SeldonDeployment mlDep = dBuilder.build();
		deploymentStore.deploymentAdded(mlDep);
		
		CloseableHttpClient mockHttpClient = Mockito.mock(CloseableHttpClient.class);
		CloseableHttpResponse mockHttpResponse = Mockito.mock(CloseableHttpResponse.class);
		Mockito.when(mockHttpResponse.getStatusLine()).thenReturn(new MockStatusLine(200));
		Mockito.when(mockHttpResponse.getEntity()).thenReturn(new MockHttpEntity("bad_json"));
		Mockito.when(mockHttpClient.execute(Mockito.any(HttpPost.class), Mockito.any(HttpContext.class))).thenReturn(mockHttpResponse);
		
		internalPredictionService.setHttpClient(mockHttpClient);
		
		Principal mockPrincipal = Mockito.mock(Principal.class);
	    Mockito.when(mockPrincipal.getName()).thenReturn(mlDep.getSpec().getOauthKey());
		
		MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
				.principal(mockPrincipal)
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(400, res.getResponse().getStatus());
	}
	
	@Test
	public void predictTest() throws Exception
	{
		final String predictJson = "{" +
         	    "\"data\": {" + 
         	    "\"ndarray\": [[1.0]]}" +
         		"}";
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
		SeldonDeployment.Builder dBuilder = SeldonDeployment.newBuilder();
		updateMessageBuilderFromJson(dBuilder, jsonStr);
		SeldonDeployment mlDep = dBuilder.build();
		deploymentStore.deploymentAdded(mlDep);
		
		CloseableHttpClient mockHttpClient = Mockito.mock(CloseableHttpClient.class);
		CloseableHttpResponse mockHttpResponse = Mockito.mock(CloseableHttpResponse.class);
		Mockito.when(mockHttpResponse.getStatusLine()).thenReturn(new MockStatusLine(200));
		Mockito.when(mockHttpResponse.getEntity()).thenReturn(new MockHttpEntity(predictJson));
		Mockito.when(mockHttpClient.execute(Mockito.any(HttpPost.class), Mockito.any(HttpContext.class))).thenReturn(mockHttpResponse);
		
		internalPredictionService.setHttpClient(mockHttpClient);
		
		Principal mockPrincipal = Mockito.mock(Principal.class);
	    Mockito.when(mockPrincipal.getName()).thenReturn(mlDep.getSpec().getOauthKey());
		
		MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/predictions")
				.principal(mockPrincipal)
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
	}
	
	@Test
	public void feedbackTest() throws Exception
	{
		final String predictJson = "{" +
         	    "request :{"+
				"\"data\": {" + 
         	    "\"ndarray\": [[1.0]]}" +
         		"}}";
		
		String jsonStr = readFile("src/test/resources/model_simple.json",StandardCharsets.UTF_8);
		SeldonDeployment.Builder dBuilder = SeldonDeployment.newBuilder();
		updateMessageBuilderFromJson(dBuilder, jsonStr);
		SeldonDeployment mlDep = dBuilder.build();
		deploymentStore.deploymentAdded(mlDep);
		
		CloseableHttpClient mockHttpClient = Mockito.mock(CloseableHttpClient.class);
		CloseableHttpResponse mockHttpResponse = Mockito.mock(CloseableHttpResponse.class);
		Mockito.when(mockHttpResponse.getStatusLine()).thenReturn(new MockStatusLine(200));
		Mockito.when(mockHttpResponse.getEntity()).thenReturn(new MockHttpEntity(predictJson));
		Mockito.when(mockHttpClient.execute(Mockito.any(HttpPost.class), Mockito.any(HttpContext.class))).thenReturn(mockHttpResponse);
		
		internalPredictionService.setHttpClient(mockHttpClient);
		
		Principal mockPrincipal = Mockito.mock(Principal.class);
	    Mockito.when(mockPrincipal.getName()).thenReturn(mlDep.getSpec().getOauthKey());
		
		MvcResult res = mvc.perform(MockMvcRequestBuilders.post("/api/v0.1/feedback")
				.principal(mockPrincipal)
    			.accept(MediaType.APPLICATION_JSON_UTF8)
    			.content(predictJson)
    			.contentType(MediaType.APPLICATION_JSON_UTF8)).andReturn();
    	String response = res.getResponse().getContentAsString();
    	System.out.println(response);
    	Assert.assertEquals(200, res.getResponse().getStatus());
	}
  
}
