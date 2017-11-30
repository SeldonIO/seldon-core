package io.seldon.engine.metrics;

import io.micrometer.spring.web.client.RestTemplateExchangeTagsProvider;

//@Configuration
public class MetricsConfig {
	
	//@Bean
	RestTemplateExchangeTagsProvider getSeldonTagProvider()
	{
		return new SeldonRestTemplateExchangeTagsProvider();
	}

}
