package io.seldon.engine.tracing;

import io.jaegertracing.Configuration;
import io.opentracing.Tracer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class TracingProvider {

  private static final Logger logger = LoggerFactory.getLogger(TracingProvider.class);

  private final Tracer tracer;
  private final boolean active;

  public TracingProvider() {
    String tracingEnv = System.getenv().get("TRACING");
    if (tracingEnv != null && "1".equals(tracingEnv)) {
      logger.info("Activating tracing");
      active = true;
      tracer = Configuration.fromEnv("seldon-svc-orch").getTracer();
    } else {
      logger.info("Not activating tracing");
      active = false;
      tracer = null;
    }
  }

  public boolean isActive() {
    return active;
  }

  public Tracer getTracer() {
    return this.tracer;
  }
}
