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
package io.seldon.wrapper.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.server.WebServerFactoryCustomizer;
import org.springframework.boot.web.servlet.server.ConfigurableServletWebServerFactory;

/**
 * Customization of the Tomcat embedded servlet engine.
 *
 * @author clive
 */
public class CustomizationBean
    implements WebServerFactoryCustomizer<ConfigurableServletWebServerFactory> {

  private static final Logger logger = LoggerFactory.getLogger(CustomizationBean.class);

  @Value("${server.port}")
  private Integer defaultServerPort;

  @Override
  public void customize(ConfigurableServletWebServerFactory container) {
    logger.info("Customizing EmbeddedServlet");

    Integer serverPort;
    serverPort = defaultServerPort;

    logger.info("setting serverPort[{}]", serverPort);
    container.setPort(serverPort);
  }
}
