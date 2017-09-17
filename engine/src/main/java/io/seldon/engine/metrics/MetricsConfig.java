package io.seldon.engine.metrics;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.micrometer.spring.web.client.RestTemplateExchangeTagsProvider;

@Configuration
public class MetricsConfig {
	
	@Bean
	RestTemplateExchangeTagsProvider getSeldonTagProvider()
	{
		return new SeldonRestTemplateExchangeTagsProvider();
	}

}
