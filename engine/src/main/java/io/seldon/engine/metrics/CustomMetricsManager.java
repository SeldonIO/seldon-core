package io.seldon.engine.metrics;

import java.util.concurrent.ConcurrentHashMap;

import org.springframework.stereotype.Component;

import com.google.common.util.concurrent.AtomicDouble;

import io.micrometer.core.instrument.Metrics;
import io.seldon.engine.predictors.PredictiveUnitState;
import io.seldon.protos.PredictionProtos.Metric;

@Component
public class CustomMetricsManager {

	private ConcurrentHashMap<PredictiveUnitState,AtomicDouble> gauges = new ConcurrentHashMap<>();
	
	public AtomicDouble get(PredictiveUnitState state,Metric metric)
	{
		if (gauges.containsKey(state))
			return gauges.get(state);
		else
		{
			AtomicDouble d = new AtomicDouble();
			gauges.put(state, d);
			Metrics.globalRegistry.gauge(metric.getKey(), d);
			return d;
		}
	}
	
	
}
