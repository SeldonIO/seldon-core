package io.seldon.engine.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.embedded.ConfigurableEmbeddedServletContainer;
import org.springframework.boot.context.embedded.EmbeddedServletContainerCustomizer;

public class CustomizationBean implements EmbeddedServletContainerCustomizer {

    private final static Logger logger = LoggerFactory.getLogger(CustomizationBean.class);
    private final static String ENGINE_SERVER_PORT_KEY = "ENGINE_SERVER_PORT";

    @Value("${server.port}")
    private Integer defaultServerPort;

    @Override
    public void customize(ConfigurableEmbeddedServletContainer container) {
        logger.info("Customizing EmbeddedServlet");

        Integer serverPort;
        { // setup the server port using the env vars
            String engineServerPortString = System.getenv().get(ENGINE_SERVER_PORT_KEY);
            if (engineServerPortString == null) {
                logger.error("FAILED to find env var [{}], will use defaults for engine server port", ENGINE_SERVER_PORT_KEY);
                serverPort = defaultServerPort;
            } else {
                logger.info("FOUND env var [{}], will use for engine server port", ENGINE_SERVER_PORT_KEY);
                serverPort = Integer.parseInt(engineServerPortString);
            }
        }

        logger.info("setting serverPort[{}]", serverPort);
        container.setPort(serverPort);
    }

}
