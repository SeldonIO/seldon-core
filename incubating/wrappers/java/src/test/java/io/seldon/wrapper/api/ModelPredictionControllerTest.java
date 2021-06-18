package io.seldon.wrapper.api;

import static io.seldon.wrapper.util.TestUtils.readFile;
import static org.hamcrest.Matchers.*;
import static org.junit.Assert.assertNotNull;
import static org.springframework.test.web.servlet.result.MockMvcResultHandlers.print;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import java.nio.charset.StandardCharsets;
import org.junit.Assert;
import org.junit.Before;
import org.junit.FixMethodOrder;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.MethodSorters;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.boot.web.server.LocalServerPort;
import org.springframework.http.MediaType;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@FixMethodOrder(MethodSorters.NAME_ASCENDING)
// @AutoConfigureMockMvc
public class ModelPredictionControllerTest {

  @Autowired private WebApplicationContext context;

  @Autowired
  ModelPredictionController modelPredictionController;

  // @Autowired
  private MockMvc mvc;

  @Before
  public void setup() {
    mvc = MockMvcBuilders.webAppContextSetup(context).build();
  }

  @LocalServerPort private int port;

  @Test
  public void testPredictLegacyGetQuery() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
                MockMvcRequestBuilders.get("/predict")
                    .accept(MediaType.APPLICATION_JSON_UTF8)
                    .param("json", predictJson)
                    .contentType(MediaType.APPLICATION_JSON_UTF8))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());
  }

  @Test
  public void testPredictLegacyPostQuery() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON_UTF8)
                .queryParam("json", predictJson)
                .contentType(MediaType.APPLICATION_JSON_UTF8))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());
  }

  @Test
  public void testPredictLegacyPostForm() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON_UTF8)
                .param("json", predictJson)
                .contentType(MediaType.APPLICATION_FORM_URLENCODED))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());
  }

  @Test
  public void testPredictLegacyButNotPredict() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON_UTF8)
                .param("json", predictJson)
                .content(predictJson)
                .contentType(MediaType.APPLICATION_FORM_URLENCODED))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());

    // if we get back a header of "application/json;charset=UTF-8" then we are hitting the  legacy predict
    Assert.assertEquals(res.getResponse().getContentType(), MediaType.APPLICATION_JSON_UTF8_VALUE);
  }

  @Test
  public void testPredictButNotPredictLegacy() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON)
                .content(predictJson)
                .contentType(MediaType.APPLICATION_JSON_UTF8))
            .andReturn();
    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());

    // if we get back a header of "application/json" then we are hitting the  legacy predict
    Assert.assertEquals(res.getResponse().getContentType(), MediaType.APPLICATION_JSON_VALUE);
  }

  @Test
  public void testPredict() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON)
                .content(predictJson)
                .contentType(MediaType.APPLICATION_JSON))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());
  }

  @Test
  public void testPredictWithUTF8Header() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON)
                .content(predictJson)
                .contentType(MediaType.APPLICATION_JSON_UTF8))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());

    // if we get back a header of "application/json" then we are hitting the  legacy predict
    Assert.assertEquals(res.getResponse().getContentType(), MediaType.APPLICATION_JSON_VALUE);
  }

  @Test
  public void testPredictWithDefaultData() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON)
                .content(predictJson)
                .contentType(MediaType.APPLICATION_JSON_UTF8)
            )
            .andDo(print())
            .andExpect(status().isOk())
            .andExpect(jsonPath("$.data", is(notNullValue())))
            .andReturn();

    // if we get back a header of "application/json" then we are hitting the  legacy predict
    Assert.assertEquals(res.getResponse().getContentType(), MediaType.APPLICATION_JSON_VALUE);
  }

  @Test
  public void testPredictWithJsonData_UTF8Header() throws Exception {
    final String predictJson = TestMessages.JSON_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
            MockMvcRequestBuilders.post("/predict")
                .accept(MediaType.APPLICATION_JSON)
                .content(predictJson)
                .contentType(MediaType.APPLICATION_JSON_UTF8))
            .andExpect(status().isOk())
            .andExpect(jsonPath("$.jsonData", is(notNullValue())))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);

    // if we get back a header of "application/json" then we are hitting the  legacy predict
    Assert.assertEquals(res.getResponse().getContentType(), MediaType.APPLICATION_JSON_VALUE);
  }

  @Test
  public void testFeedback() throws Exception {
    final String predictJson = TestMessages.DEFAULT_DATA;
    assertNotNull(predictJson);

    MvcResult res =
        mvc.perform(
                MockMvcRequestBuilders.get("/send-feedback")
                    .accept(MediaType.APPLICATION_JSON_UTF8)
                    .param("json", predictJson)
                    .contentType(MediaType.APPLICATION_JSON_UTF8))
            .andReturn();

    String response = res.getResponse().getContentAsString();
    System.out.println(response);
    Assert.assertEquals(200, res.getResponse().getStatus());
  }
}
