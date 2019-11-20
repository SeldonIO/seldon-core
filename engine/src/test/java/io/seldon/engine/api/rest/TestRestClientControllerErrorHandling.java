package io.seldon.engine.api.rest;

import io.seldon.engine.filters.XSSFilter;
import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.service.InternalPredictionService;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Matchers;
import org.mockito.Mockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.boot.web.server.LocalServerPort;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.HttpStatusCodeException;
import org.springframework.web.context.WebApplicationContext;

import java.net.URI;
import java.nio.charset.StandardCharsets;

import static io.seldon.engine.util.TestUtils.readFile;

@RunWith(SpringRunner.class)
@ActiveProfiles("test")
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
public class TestRestClientControllerErrorHandling
{

    @Autowired
    private WebApplicationContext context;

    @Autowired private EnginePredictor enginePredictor;

    // @Autowired
    private MockMvc mvc;

    @Autowired private RestClientController restController;

    @Before
    public void setup() throws Exception {

        mvc = MockMvcBuilders.webAppContextSetup(context).addFilters(new XSSFilter()).build();
    }

    @LocalServerPort
    private int port;

    @Autowired private TestRestTemplate testRestErrorTemplate;

    @Autowired private InternalPredictionService internalPredictionService;

    @Test
    public void testModelPredictionNon200Response() throws Exception {
        String jsonStr = readFile("src/test/resources/model_simple.json", StandardCharsets.UTF_8);
        String responseStr =
                readFile("src/test/resources/response_status.json", StandardCharsets.UTF_8);
        io.seldon.protos.DeploymentProtos.PredictorSpec.Builder PredictorSpecBuilder = io.seldon.protos.DeploymentProtos.PredictorSpec.newBuilder();
        EnginePredictor.updateMessageBuilderFromJson(PredictorSpecBuilder, jsonStr);
        io.seldon.protos.DeploymentProtos.PredictorSpec predictorSpec = PredictorSpecBuilder.build();
        final String predictJson = "{" + "\"binData\": \"MTIz\"" + "}";
        ReflectionTestUtils.setField(enginePredictor, "predictorSpec", predictorSpec);

        HttpStatusCodeException exception = HttpServerErrorException.InternalServerError
                .create(HttpStatus.BAD_REQUEST, "status text", HttpHeaders.EMPTY, responseStr.getBytes(StandardCharsets.UTF_8), StandardCharsets.UTF_8);

        Mockito.when(
                testRestErrorTemplate
                        .getRestTemplate()
                        .postForEntity(
                                Matchers.<URI>any(),
                                Matchers.<HttpEntity<MultiValueMap<String, String>>>any(),
                                Matchers.<Class<String>>any()))
                .thenThrow(exception);

        MvcResult res =
                mvc.perform(
                        MockMvcRequestBuilders.post("/api/v0.1/predictions")
                                .accept(MediaType.APPLICATION_JSON_UTF8)
                                .content(predictJson)
                                .contentType(MediaType.APPLICATION_JSON_UTF8))
                        .andReturn();

        // Check for returned response that wraps the ApiException into SeldonMessage
        Assert.assertEquals(200, res.getResponse().getStatus());
        Assert.assertEquals(responseStr, res.getResponse().getContentAsString());
    }
}
