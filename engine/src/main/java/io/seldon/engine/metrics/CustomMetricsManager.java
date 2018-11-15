package io.seldon.engine.metrics;

import java.util.concurrent.ConcurrentHashMap;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import com.google.common.util.concurrent.AtomicDouble;

import io.micrometer.core.instrument.Meter;
import io.micrometer.core.instrument.Metrics;
import io.micrometer.core.instrument.Tag;
import io.seldon.protos.PredictionProtos.Metric;

/**
 * 
 * @author clive
 * Handles the storage of gauges for custom metrics.
 *
 */
@Component
public class CustomMetricsManager {

	private final static Logger logger = LoggerFactory.getLogger(CustomMetricsManager.class);
	private ConcurrentHashMap<Meter.Id,AtomicDouble> gauges = new ConcurrentHashMap<>();
	

	public AtomicDouble get(Iterable<Tag> tags,Metric metric)
	{
		Meter.Id id = Metrics.globalRegistry.createId(metric.getKey(), tags, "");
		if (gauges.containsKey(id))
			return gauges.get(id);
		else
		{
			logger.info("Creating new metric Id for {}",metric.toString());
			AtomicDouble d = new AtomicDouble();
			gauges.put(id, d);
			Metrics.globalRegistry.gauge(metric.getKey(), tags, d);
			return d;
		}
	}
	
	
}
