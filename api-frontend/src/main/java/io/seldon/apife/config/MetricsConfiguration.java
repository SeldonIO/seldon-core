package io.seldon.apife.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.micrometer.spring.web.servlet.WebMvcTagsProvider;
import io.seldon.apife.metrics.AuthorizedWebMvcTagsProvider;

@Configuration
public class MetricsConfiguration {

	@Bean
	WebMvcTagsProvider getWebmvcTagConfigurer()
	{
		return new AuthorizedWebMvcTagsProvider();
	}
	
}
