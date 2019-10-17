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

import io.seldon.engine.predictors.EnginePredictor;
import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.web.server.WebServerFactoryCustomizer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

public class AppConfig {

  @Bean
  @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
  public WebServerFactoryCustomizer containerCustomizer() {
    return new CustomizationBean();
  }

  @Bean(initMethod = "init", destroyMethod = "cleanup")
  @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
  public EnginePredictor enginePredictor() {
    return new EnginePredictor();
  }
}
