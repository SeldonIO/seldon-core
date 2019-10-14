/**
 * ***************************************************************************** Copyright 2017
 * Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * <p>Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 *
 * <p>http://www.apache.org/licenses/LICENSE-2.0
 *
 * <p>Unless required by applicable law or agreed to in writing, software distributed under the
 * License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 * *****************************************************************************
 */
package io.seldon.engine.config;

import org.apache.http.client.HttpClient;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.HttpComponentsClientHttpRequestFactory;
import org.springframework.web.client.RestTemplate;

@Configuration
public class RestTemplateConfig {

  private static final int DEFAULT_REQ_TIMEOUT = 1000;
  private static final int DEFAULT_CON_TIMEOUT = 1000;
  private static final int DEFAULT_SOCKET_TIMEOUT = 2000;
  private static final int DEFAULT_POOL_MAX_TOTAL = 150;
  private static final int DEFAULT_POOL_MAX_PER_ROUTE = 25;

  @Bean
  public PoolingHttpClientConnectionManager poolingHttpClientConnectionManager() {
    PoolingHttpClientConnectionManager result = new PoolingHttpClientConnectionManager();
    result.setMaxTotal(DEFAULT_POOL_MAX_TOTAL);
    result.setDefaultMaxPerRoute(DEFAULT_POOL_MAX_PER_ROUTE);
    return result;
  }

  @Bean
  public RequestConfig requestConfig() {
    RequestConfig result =
        RequestConfig.custom()
            .setConnectionRequestTimeout(DEFAULT_REQ_TIMEOUT)
            .setConnectTimeout(DEFAULT_CON_TIMEOUT)
            .setSocketTimeout(DEFAULT_SOCKET_TIMEOUT)
            .build();
    return result;
  }

  @Bean
  public CloseableHttpClient httpClient(
      PoolingHttpClientConnectionManager poolingHttpClientConnectionManager,
      RequestConfig requestConfig) {
    CloseableHttpClient result =
        HttpClientBuilder.create()
            .setConnectionManager(poolingHttpClientConnectionManager)
            .setDefaultRequestConfig(requestConfig)
            .build();
    return result;
  }

  @Bean
  public RestTemplate restTemplate(HttpClient httpClient) {
    HttpComponentsClientHttpRequestFactory requestFactory =
        new HttpComponentsClientHttpRequestFactory();
    requestFactory.setHttpClient(httpClient);
    return new RestTemplate(requestFactory);
  }
}
