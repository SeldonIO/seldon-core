package io.seldon.engine.tracing;

import org.springframework.stereotype.Component;

import io.jaegertracing.Configuration;
import io.opentracing.Tracer;

@Component
public class TracingProvider {

	private final Tracer tracer;
	private final boolean active;
	
	
	public TracingProvider()
	{
		String tracingEnv = System.getenv().get("TRACING");
		if (tracingEnv != null)
		{
			active = true;
			tracer = Configuration.fromEnv("seldon-svc-orch").getTracer();
		}
		else
		{
			active = false;
			tracer = null;
		}
	}
	
	
	public boolean isActive() {
		return active;
	}



	public Tracer getTracer()
	{
		return this.tracer;
	}
	
}
