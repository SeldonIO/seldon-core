package io.seldon.engine.metrics;

import com.google.common.util.concurrent.AtomicDouble;
import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Meter;
import io.micrometer.core.instrument.Meter.Type;
import io.micrometer.core.instrument.Metrics;
import io.micrometer.core.instrument.Tag;
import io.micrometer.core.instrument.Tags;
import io.micrometer.core.instrument.Timer;
import io.seldon.protos.PredictionProtos.Metric;
import java.util.concurrent.ConcurrentHashMap;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/** @author clive Handles the storage of gauges for custom metrics. */
@Component
public class CustomMetricsManager {

  private static final Logger logger = LoggerFactory.getLogger(CustomMetricsManager.class);
  private ConcurrentHashMap<Meter.Id, AtomicDouble> gauges = new ConcurrentHashMap<>();

  public AtomicDouble getGaugeValue(Iterable<Tag> tags, Metric metric) {
    Tags tagsObj = Tags.concat(tags);
    Meter.Id id = new Meter.Id(metric.getKey(), tagsObj, "", "", Type.GAUGE);
    if (gauges.containsKey(id)) return gauges.get(id);
    else {
      logger.info("Creating new metric Id for {}", metric.toString());
      try {
        AtomicDouble d = new AtomicDouble();
        gauges.put(id, d);
        Metrics.gauge(metric.getKey(), tags, d);
        return d;
      } catch (IllegalArgumentException e) {
        logger.warn(
            "Can't create gauge Metric. Probably same name exists with different number of tags. Not allowed in Prometheus Registry. Key {}",
            metric.getKey(),
            e);
        throw e;
      }
    }
  }

  public Counter getCounter(Iterable<Tag> tags, Metric metric) {
    try {
      Counter counter = Metrics.counter(metric.getKey(), tags);
      return counter;
    } catch (IllegalArgumentException e) {
      logger.warn(
          "Can't create counter Metric. Probably same name exists with different number of tags. Not allowed in Prometheus Registry. Key {}",
          metric.getKey(),
          e);
      throw e;
    }
  }

  public Timer getTimer(Iterable<Tag> tags, Metric metric) {
    try {
      Timer timer = Metrics.timer(metric.getKey(), tags);
      return timer;
    } catch (IllegalArgumentException e) {
      logger.warn(
          "Can't create Timer Metric. Probably same name exists with different number of tags. Not allowed in Prometheus Registry. Key: {}",
          metric.getKey(),
          e);
      throw e;
    }
  }
}
