/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.servlet.server.ConfigurableServletWebServerFactory;
import org.springframework.boot.web.server.WebServerFactoryCustomizer;

public class CustomizationBean implements WebServerFactoryCustomizer<ConfigurableServletWebServerFactory> {

    private final static Logger logger = LoggerFactory.getLogger(CustomizationBean.class);
    private final static String ENGINE_SERVER_PORT_KEY = "ENGINE_SERVER_PORT";

    @Value("${server.port}")
    private Integer defaultServerPort;

    @Override
    public void customize(ConfigurableServletWebServerFactory factory) {
        logger.info("Customizing EmbeddedServlet");

        Integer serverPort;
        { // setup the server port using the env vars
            String engineServerPortString = System.getenv().get(ENGINE_SERVER_PORT_KEY);
            if (engineServerPortString == null) {
                logger.warn("FAILED to find env var [{}], will use defaults for engine server port", ENGINE_SERVER_PORT_KEY);
                serverPort = defaultServerPort;
            } else {
                logger.info("FOUND env var [{}], will use for engine server port", ENGINE_SERVER_PORT_KEY);
                serverPort = Integer.parseInt(engineServerPortString);
            }
        }

        logger.info("setting serverPort[{}]", serverPort);
        factory.setPort(serverPort);
    }

}
