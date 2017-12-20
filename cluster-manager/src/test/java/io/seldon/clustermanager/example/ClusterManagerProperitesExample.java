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
package io.seldon.clustermanager.example;

import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;

import io.seldon.clustermanager.ClusterManagerProperites;

public class ClusterManagerProperitesExample {

    @EnableAutoConfiguration
    public static class AppConfig {

        @ConfigurationProperties(prefix = "io.seldon.clustermanager")
        @Bean
        public ClusterManagerProperites clusterManagerProperites() {
            return new ClusterManagerProperites();
        }
    }

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        ClusterManagerProperites clusterManagerProperites = ctx.getBean(ClusterManagerProperites.class);
        System.out.println(clusterManagerProperites);

        ctx.close();

    }
}
