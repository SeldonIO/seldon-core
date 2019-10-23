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

import com.google.common.collect.Sets;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import org.apache.catalina.connector.Connector;
import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.boot.web.servlet.server.ServletWebServerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class TomcatConfig {

  @Value("${server.port}")
  private String serverPort;

  @Value("${management.port:${server.port}}")
  private String managementPort;

  @Value("${server.additionalPorts:}")
  private String additionalPorts;

  @Bean
  public ServletWebServerFactory servletContainer() {
    TomcatServletWebServerFactory tomcat = new TomcatServletWebServerFactory();
    Connector[] additionalConnectors = this.additionalConnector();
    if (additionalConnectors != null && additionalConnectors.length > 0) {
      tomcat.addAdditionalTomcatConnectors(additionalConnectors);
    }

    return tomcat;
  }

  private Connector[] additionalConnector() {
    if (StringUtils.isBlank(this.additionalPorts)) {
      return null;
    }

    Set<String> defaultPorts = Sets.newHashSet(this.serverPort, this.managementPort);
    String[] ports = this.additionalPorts.split(",");
    List<Connector> result = new ArrayList<>();
    for (String port : ports) {
      if (!defaultPorts.contains(port)) {
        Connector connector = new Connector("org.apache.coyote.http11.Http11NioProtocol");
        connector.setScheme("http");
        connector.setPort(Integer.valueOf(port));
        result.add(connector);
      }
    }

    return result.toArray(new Connector[] {});
  }
}
