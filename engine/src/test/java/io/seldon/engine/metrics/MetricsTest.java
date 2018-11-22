package io.seldon.engine.metrics;

import java.util.Arrays;

import org.junit.Assert;
import org.junit.Test;

import com.google.common.util.concurrent.AtomicDouble;

import io.micrometer.core.instrument.Tag;
import io.seldon.engine.metrics.CustomMetricsManager;
import io.seldon.protos.PredictionProtos.Metric;
import io.seldon.protos.PredictionProtos.Metric.MetricType;


public class MetricsTest {

	@Test
	public void testGauge()
	{
		CustomMetricsManager m = new CustomMetricsManager();
		Iterable<Tag> tags = Arrays.asList(Tag.of("tag1", "tag1value1"));
		Metric metric1 = Metric.newBuilder().setKey("key1").setValue(1.0f).setType(MetricType.GAUGE).build();
		AtomicDouble v1 = m.get(tags, metric1);
		v1.set(metric1.getValue());
		Metric metric2 = Metric.newBuilder().setKey("key1").setValue(2.0f).setType(MetricType.GAUGE).build();
		AtomicDouble v2 = m.get(tags, metric2);
		Assert.assertEquals(1.0D, v2.get(),0.01);
		Assert.assertEquals(v1, v2);
		Metric metric3 = Metric.newBuilder().setKey("key2").setValue(2.0f).setType(MetricType.GAUGE).build();
		AtomicDouble v3 = m.get(tags, metric3);
		Assert.assertNotEquals(v1, v3);
		v2.set(metric2.getValue());
		AtomicDouble v4 = m.get(tags, metric1);
		Assert.assertEquals(2.0D, v4.get(),0.01);
	}
	
}
