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
package io.seldon.apife.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.connection.jedis.JedisConnectionFactory;
import org.springframework.data.redis.core.StringRedisTemplate;

import redis.clients.jedis.JedisPoolConfig;

@Configuration
public class RedisConfig {

    private final static Logger logger = LoggerFactory.getLogger(RedisConfig.class);
    private final static String SELDON_CLUSTER_MANAGER_REDIS_HOST_KEY = "SELDON_CLUSTER_MANAGER_REDIS_HOST";

    @Bean
    public RedisConnectionFactory redisConnectionFactory() {
        String default_host = "localhost";
        int port = 6379;
        String host = null;
        { // setup the host using the env vars
            host = System.getenv().get(SELDON_CLUSTER_MANAGER_REDIS_HOST_KEY);
            if (host == null) {
                logger.error(String.format("FAILED to find env var [%s]", SELDON_CLUSTER_MANAGER_REDIS_HOST_KEY));
                host = default_host;
            }
        }

        logger.info(String.format("setting up connection factory, host[%s] port[%d]", host, port));
        JedisConnectionFactory jedisConnectionFactory = new JedisConnectionFactory();
        jedisConnectionFactory.setHostName(host);
        jedisConnectionFactory.setPort(port);
        jedisConnectionFactory.setPassword("");
        jedisConnectionFactory.setUsePool(true);
        jedisConnectionFactory.setPoolConfig(new JedisPoolConfig());
        return jedisConnectionFactory;
    }

    @Bean
    StringRedisTemplate template(RedisConnectionFactory connectionFactory) {
        return new StringRedisTemplate(connectionFactory);
    }
}
