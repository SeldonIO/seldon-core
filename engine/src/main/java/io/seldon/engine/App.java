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
package io.seldon.engine;

import java.util.concurrent.Executor;
import java.util.concurrent.TimeUnit;

import org.apache.catalina.connector.Connector;
import org.apache.tomcat.util.threads.ThreadPoolExecutor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.embedded.ConfigurableEmbeddedServletContainer;
import org.springframework.boot.context.embedded.EmbeddedServletContainerCustomizer;
import org.springframework.boot.context.embedded.tomcat.TomcatConnectorCustomizer;
import org.springframework.boot.context.embedded.tomcat.TomcatEmbeddedServletContainerFactory;
import org.springframework.context.ApplicationListener;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Import;
import org.springframework.context.event.ContextClosedEvent;
import org.springframework.scheduling.annotation.EnableAsync;

import io.seldon.engine.config.AppConfig;

@SpringBootApplication
@EnableAsync
@Import({ AppConfig.class })
public class App {
    public static void main(String[] args) throws Exception {
        SpringApplication.run(App.class, args);
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
