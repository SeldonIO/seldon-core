package io.seldon.wrapper.api;

import static io.seldon.wrapper.util.TestUtils.readFileFromAbsolutePathOrResources;

final public class TestMessages {

  /**
   * All possible fields based on the SeldonMessage Proto:
   *   https://docs.seldon.io/projects/seldon-core/en/v1.6.0/reference/apis/prediction.html
   */
  public static final String TF_DATA = readFile("requests/defaultData.json");
  public static final String DEFAULT_DATA = TF_DATA;
  public static final String JSON_DATA = readFile("requests/jsonData.json");
  // TODO: add binData
  // TODO: add strData
  // TODO: add customData

  private static String readFile(String file) {
    return readFileFromAbsolutePathOrResources(file);
  }
}
