package io.seldon.engine.config;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.context.embedded.EmbeddedServletContainerCustomizer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.engine.predictors.EnginePredictor;

public class AppConfig {

    @Bean
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public EmbeddedServletContainerCustomizer containerCustomizer() {
        return new CustomizationBean();
    }

    @Bean(initMethod = "init", destroyMethod = "cleanup")
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public EnginePredictor enginePredictor() {
        return new EnginePredictor();
    }
}
