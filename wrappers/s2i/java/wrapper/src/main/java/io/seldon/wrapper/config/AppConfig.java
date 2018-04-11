package io.seldon.wrapper.config;

import java.util.concurrent.Executor;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

import org.apache.catalina.connector.Connector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.context.embedded.ConfigurableEmbeddedServletContainer;
import org.springframework.boot.context.embedded.EmbeddedServletContainerCustomizer;
import org.springframework.boot.context.embedded.tomcat.TomcatConnectorCustomizer;
import org.springframework.boot.context.embedded.tomcat.TomcatEmbeddedServletContainerFactory;
import org.springframework.context.ApplicationListener;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;
import org.springframework.context.event.ContextClosedEvent;


/**
 * Spring REST config 
 * @author clive
 *
 */
public class AppConfig {

    @Bean
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public EmbeddedServletContainerCustomizer containerCustomizer() {
        return new CustomizationBean();
    }

    @Bean
    public GracefulShutdown gracefulShutdown() {
        return new GracefulShutdown();
    }
    
    @Bean
    public EmbeddedServletContainerCustomizer tomcatCustomizer() {
        return new EmbeddedServletContainerCustomizer() {

            @Override
            public void customize(ConfigurableEmbeddedServletContainer container) {
                if (container instanceof TomcatEmbeddedServletContainerFactory) {
                    ((TomcatEmbeddedServletContainerFactory) container)
                            .addConnectorCustomizers(gracefulShutdown());
                }

            }
        };
    }

    /**
     * Ensure a graceful shutdown of Tomcat to allow requests in process to complete.
     * @author clive
     *
     */
    private static class GracefulShutdown implements TomcatConnectorCustomizer,
            ApplicationListener<ContextClosedEvent> {

        private static final Logger log = LoggerFactory.getLogger(GracefulShutdown.class);

        private volatile Connector connector;

        @Override
        public void customize(Connector connector) {
            this.connector = connector;
        }

        @Override
        public void onApplicationEvent(ContextClosedEvent event) {
        	log.info("Starting graceful shutdown of Tomcat");
        	if (this.connector != null)
        	{
        		this.connector.pause();
                Executor executor = this.connector.getProtocolHandler().getExecutor();
                if (executor instanceof ThreadPoolExecutor) {
                    try {
                        ThreadPoolExecutor threadPoolExecutor = (ThreadPoolExecutor) executor;
                        threadPoolExecutor.shutdown();
                        if (!threadPoolExecutor.awaitTermination(20, TimeUnit.SECONDS)) {
                            log.warn("Tomcat thread pool did not shut down gracefully within "
                                    + "20 seconds. Proceeding with forceful shutdown");
                        }
                        else
                        {
                            log.info("Thread pool has closed");
                        }
                    }
                    catch (InterruptedException ex) {
                        Thread.currentThread().interrupt();
                    }
                }
        	}
        }

    }
}

