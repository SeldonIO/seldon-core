package io.seldon.engine.metrics;

import com.google.common.util.concurrent.AtomicDouble;
import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Meter;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Metrics;
import io.micrometer.core.instrument.Tag;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import io.micrometer.prometheus.PrometheusConfig;
import io.micrometer.prometheus.PrometheusMeterRegistry;
import io.seldon.protos.PredictionProtos.Metric;
import io.seldon.protos.PredictionProtos.Metric.MetricType;
import java.util.ArrayList;
import java.util.Arrays;
import org.junit.After;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;

public class MetricsTest {

  @Before
  public void cleanMicrometer() {
    for (; ; ) {
      if (Metrics.globalRegistry.getRegistries().size() > 0)
        Metrics.removeRegistry(Metrics.globalRegistry.getRegistries().iterator().next());
      else break;
    }
  }

  @After
  public void cleanMicrometerAfter() {
    for (; ; ) {
      if (Metrics.globalRegistry.getRegistries().size() > 0)
        Metrics.removeRegistry(Metrics.globalRegistry.getRegistries().iterator().next());
      else break;
    }
  }

  @Test
  public synchronized void testGauge() {
    SimpleMeterRegistry reg = new SimpleMeterRegistry();
    Metrics.addRegistry(reg);
    CustomMetricsManager m = new CustomMetricsManager();
    Iterable<Tag> tags = Arrays.asList(Tag.of("gtag1", "tag1value1"));
    Metric metric1 =
        Metric.newBuilder().setKey("gkey1").setValue(1.0f).setType(MetricType.GAUGE).build();
    AtomicDouble v1 = m.getGaugeValue(tags, metric1);
    v1.set(metric1.getValue());
    Metric metric2 =
        Metric.newBuilder().setKey("gkey1").setValue(2.0f).setType(MetricType.GAUGE).build();
    AtomicDouble v2 = m.getGaugeValue(tags, metric2);
    Assert.assertEquals(1.0D, v2.get(), 0.01);
    Assert.assertEquals(v1, v2);
    Metric metric3 =
        Metric.newBuilder().setKey("gkey2").setValue(2.0f).setType(MetricType.GAUGE).build();
    AtomicDouble v3 = m.getGaugeValue(tags, metric3);
    Assert.assertNotEquals(v1, v3);
    v2.set(metric2.getValue());
    AtomicDouble v4 = m.getGaugeValue(tags, metric1);
    Assert.assertEquals(2.0D, v4.get(), 0.01);
    Metrics.removeRegistry(reg);
  }

  @Test
  public void testCounter() {
    SimpleMeterRegistry reg = new SimpleMeterRegistry();
    Metrics.addRegistry(reg);
    CustomMetricsManager m = new CustomMetricsManager();
    Iterable<Tag> tags = Arrays.asList(Tag.of("tag1", "tag1value1"));
    Metric metric1 =
        Metric.newBuilder().setKey("ckey3").setValue(1.0f).setType(MetricType.COUNTER).build();
    Counter c1 = m.getCounter(tags, metric1);
    c1.increment(1);
    Metric metric2 =
        Metric.newBuilder().setKey("ckey3").setValue(2.0f).setType(MetricType.COUNTER).build();
    Counter c2 = m.getCounter(tags, metric1);
    Assert.assertEquals(c1, c2);
    double d1 = c2.count();
    Assert.assertEquals(1.0D, c2.count(), 0.01);
    Metric metric3 =
        Metric.newBuilder().setKey("ckey4").setValue(2.0f).setType(MetricType.COUNTER).build();
    Counter c3 = m.getCounter(tags, metric3);
    Assert.assertNotEquals(c1, c3);
    c2.increment();
    Counter c4 = m.getCounter(tags, metric1);
    Assert.assertEquals(2.0D, c4.count(), 0.01);
    Metrics.removeRegistry(reg);
  }

  @Test(expected = IllegalArgumentException.class)
  public synchronized void testLabelsExceptionLowLevel() {
    PrometheusMeterRegistry reg = new PrometheusMeterRegistry(PrometheusConfig.DEFAULT);
    Metrics.addRegistry(reg);
    Iterable<Tag> tags1 = Arrays.asList(Tag.of("tag1", "tag1value1"));
    Iterable<Tag> tags2 = Arrays.asList(Tag.of("tag2", "tag2value1"));
    Metric metric1 =
        Metric.newBuilder().setKey("ckey1").setValue(1.0f).setType(MetricType.COUNTER).build();
    Counter counter1 = Metrics.counter(metric1.getKey(), tags1);
    try {
      // Exception throw here as Prometheus registry can't have same id with different length tags
      Counter counter2 = Metrics.counter(metric1.getKey(), new ArrayList<>());
    } finally {
      Metrics.removeRegistry(reg);
      Meter m = Metrics.globalRegistry.remove(counter1);
    }
  }

  @Test(expected = IllegalArgumentException.class)
  public void testLabelsException() {
    for (MeterRegistry r : Metrics.globalRegistry.getRegistries()) {
      Metrics.removeRegistry(r);
    }
    PrometheusMeterRegistry reg = new PrometheusMeterRegistry(PrometheusConfig.DEFAULT);
    Metrics.addRegistry(reg);

    CustomMetricsManager m = new CustomMetricsManager();

    Iterable<Tag> tags1 = Arrays.asList(Tag.of("tag1", "tag1value1"));
    Iterable<Tag> tags2 = Arrays.asList(Tag.of("tag2", "tag2value1"));
    Metric metric1 =
        Metric.newBuilder().setKey("ckey2").setValue(1.0f).setType(MetricType.COUNTER).build();

    Counter counter1 = m.getCounter(tags1, metric1);
    try {
      // Exception throw here as Prometheus registry can't have same id with different length tags
      Counter counter2 = m.getCounter(new ArrayList<>(), metric1);
    } finally {
      Meter met = Metrics.globalRegistry.remove(counter1);
    }
  }
}
