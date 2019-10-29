package io.seldon.engine.config;

import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import io.seldon.engine.grpc.GrpcChannelHandler;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.engine.tracing.TracingProvider;
import java.io.IOException;
import java.time.Duration;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.context.annotation.Profile;
import org.springframework.http.client.ClientHttpRequestFactory;
import org.springframework.web.client.RestTemplate;

@Configuration
@Profile("test")
public class TestConfig {
  @Autowired private AnnotationsConfig annotationsConfig;

  @Autowired private TracingProvider tracingProvider;

  @Autowired private GrpcChannelHandler grpcChannelHandler;

  @Bean
  @Primary
  public RestTemplateBuilder restTemplateBuilder() {

    RestTemplateBuilder rtb = mock(RestTemplateBuilder.class);
    RestTemplate restTemplate = mock(RestTemplate.class);
    ClientHttpRequestFactory requestFactory = mock(ClientHttpRequestFactory.class);
    when(rtb.setConnectTimeout(any(Duration.class))).thenReturn(rtb);
    when(rtb.setReadTimeout(any(Duration.class))).thenReturn(rtb);
    when(rtb.build()).thenReturn(restTemplate);
    when(restTemplate.getRequestFactory()).thenReturn(requestFactory);

    return rtb;
  }

  @Bean
  @Primary
  public InternalPredictionService service() throws IOException {

    return new InternalPredictionService(
        restTemplateBuilder(), annotationsConfig, grpcChannelHandler, tracingProvider);
  }
}
