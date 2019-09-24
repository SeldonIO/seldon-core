package io.seldon.engine.config;

import io.seldon.engine.grpc.GrpcChannelHandler;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.engine.tracing.TracingProvider;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.context.annotation.Profile;
import org.springframework.web.client.RestTemplate;
import org.springframework.http.client.ClientHttpRequestFactory;

import java.io.IOException;

import static org.mockito.Matchers.anyInt;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

@Configuration
@Profile("test")
public class TestConfig {
    @Autowired
    private AnnotationsConfig annotationsConfig;

    @Autowired
    private TracingProvider tracingProvider;

    @Autowired
    private GrpcChannelHandler grpcChannelHandler;

    @Bean
    @Primary
    public RestTemplateBuilder restTemplateBuilder() {

        RestTemplateBuilder rtb = mock(RestTemplateBuilder.class);
        RestTemplate restTemplate = mock(RestTemplate.class);
        ClientHttpRequestFactory requestFactory = mock(ClientHttpRequestFactory.class);
        when(rtb.setConnectTimeout(anyInt())).thenReturn(rtb);
        when(rtb.setReadTimeout(anyInt())).thenReturn(rtb);
        when(rtb.build()).thenReturn(restTemplate);
        when(restTemplate.getRequestFactory()).thenReturn(requestFactory);

        return rtb;
    }

    @Bean
    @Primary
    public InternalPredictionService service() throws IOException {

        return new InternalPredictionService(restTemplateBuilder(), annotationsConfig, grpcChannelHandler, tracingProvider);

    }

}
