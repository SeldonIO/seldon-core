package io.seldon.apife.config;

import io.prometheus.client.exporter.MetricsServlet;
import io.prometheus.client.spring.boot.SpringBootMetricsCollector;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.actuate.endpoint.PublicMetrics;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.web.servlet.ServletRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.Collection;

@Configuration
@ConditionalOnClass(SpringBootMetricsCollector.class)
class PrometheusAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean(SpringBootMetricsCollector.class)
    SpringBootMetricsCollector springBootMetricsCollector(Collection<PublicMetrics> publicMetrics) {

        SpringBootMetricsCollector springBootMetricsCollector = new SpringBootMetricsCollector(publicMetrics);
        springBootMetricsCollector.register();

        return springBootMetricsCollector;
    }

    @Bean
    @ConditionalOnMissingBean(name = "prometheusMetricsServletRegistrationBean")
    ServletRegistrationBean prometheusMetricsServletRegistrationBean(@Value("${prometheus.metrics.path:/prometheus}") String metricsPath) {
        return new ServletRegistrationBean(new MetricsServlet(), metricsPath);
    }
}
