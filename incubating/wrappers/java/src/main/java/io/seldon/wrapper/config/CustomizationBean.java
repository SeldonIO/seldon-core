/**
 * ***************************************************************************** 
 * 
 * Copyright (c) 2024 Seldon Technologies Ltd.
 * 
 * Use of this software is governed BY
 * (1) the license included in the LICENSE file or
 * (2) if the license included in the LICENSE file is the Business Source License 1.1,
 * the Change License after the Change Date as each is defined in accordance with the LICENSE file.
 *
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
