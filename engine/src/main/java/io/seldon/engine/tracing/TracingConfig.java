package io.seldon.engine.tracing;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.jaegertracing.internal.JaegerTracer;
import io.jaegertracing.internal.reporters.InMemoryReporter;
import io.jaegertracing.internal.samplers.ConstSampler;
import io.jaegertracing.spi.Reporter;
import io.jaegertracing.spi.Sampler;

@Configuration
public class TracingConfig {

	@Autowired
	TracingProvider tracingProvider;
	
	 @Bean
	 public io.opentracing.Tracer jaegerTracer() {
		 if (!tracingProvider.isActive())
		 {
			 final Reporter reporter = new InMemoryReporter();
			 final Sampler sampler = new ConstSampler(false);
			 return new JaegerTracer.Builder("untraced-service")
					 .withReporter(reporter)
					 .withSampler(sampler)
					 .build();
		 }
		 else
			 return tracingProvider.getTracer();
	 }
}
