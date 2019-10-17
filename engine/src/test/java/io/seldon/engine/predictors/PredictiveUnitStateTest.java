package io.seldon.engine.predictors;

import io.kubernetes.client.proto.V1.Container;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import java.util.HashMap;
import java.util.Map;
import org.junit.Assert;
import org.junit.Test;

public class PredictiveUnitStateTest {

  @Test
  public void testImageExtractionBasic() {
    final String name = "test";
    PredictiveUnit pu =
        PredictiveUnit.newBuilder()
            .setName(name)
            .setEndpoint(Endpoint.newBuilder().setServiceHost("host"))
            .build();
    Map<String, Container> containersMap = new HashMap<String, Container>();
    containersMap.put(name, Container.newBuilder().setImage("myimage:0.1").build());
    PredictiveUnitState pus = new PredictiveUnitState(pu, containersMap);
    Assert.assertEquals("myimage", pus.imageName);
    Assert.assertEquals("0.1", pus.imageVersion);
  }

  @Test
  public void testImageExtractionWithPort() {
    final String name = "test";
    PredictiveUnit pu =
        PredictiveUnit.newBuilder()
            .setName(name)
            .setEndpoint(Endpoint.newBuilder().setServiceHost("host"))
            .build();
    Map<String, Container> containersMap = new HashMap<String, Container>();
    containersMap.put(
        name, Container.newBuilder().setImage("myrep:1234/someorg/myimage:0.1").build());
    PredictiveUnitState pus = new PredictiveUnitState(pu, containersMap);
    Assert.assertEquals("myrep:1234/someorg/myimage", pus.imageName);
    Assert.assertEquals("0.1", pus.imageVersion);
  }
}
