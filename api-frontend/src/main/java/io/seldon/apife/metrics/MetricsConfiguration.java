package io.seldon.apife.metrics;

import io.micrometer.spring.web.servlet.WebMvcTagsProvider;

//@Configuration
public class MetricsConfiguration {

	//@Bean
	WebMvcTagsProvider getWebmvcTagConfigurer()
	{
		return new AuthorizedWebMvcTagsProvider();
	}
	
}
