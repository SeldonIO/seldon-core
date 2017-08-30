package io.seldon.apife.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.micrometer.spring.web.WebmvcTagConfigurer;
import io.seldon.apife.metrics.AuthorizedWebmvcTagConfigurer;

@Configuration
public class MetricsConfiguration {

	@Bean
	WebmvcTagConfigurer getWebmvcTagConfigurer()
	{
		return new AuthorizedWebmvcTagConfigurer();
	}
	
}
