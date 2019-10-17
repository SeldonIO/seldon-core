package io.seldon.engine.service;

import io.seldon.protos.DeploymentProtos.Endpoint;
import org.junit.Assert;
import org.junit.Test;

public class UriCacheTest {

  @Test
  public void testUri() {
    Endpoint endpointA = Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).build();
    Endpoint endpointA2 =
        Endpoint.newBuilder().setServiceHost("hostA").setServicePort(1000).build();
    Endpoint endpointB = Endpoint.newBuilder().setServiceHost("hostB").setServicePort(1000).build();
    final String predictPath = "/predict";
    final String predictPath2 = "/predict";
    final String feedbackPath = "/feedback";

    final String key1 = InternalPredictionService.getUriKey(endpointA, predictPath);
    final String key2 = InternalPredictionService.getUriKey(endpointB, predictPath);
    final String key3 = InternalPredictionService.getUriKey(endpointA, feedbackPath);

    Assert.assertNotEquals(key1, key2);
    Assert.assertNotEquals(key1, key3);

    final String key4 = InternalPredictionService.getUriKey(endpointA2, predictPath);

    Assert.assertEquals(key1, key4);

    final String key5 = InternalPredictionService.getUriKey(endpointA, predictPath2);

    Assert.assertEquals(key1, key5);
  }
}
