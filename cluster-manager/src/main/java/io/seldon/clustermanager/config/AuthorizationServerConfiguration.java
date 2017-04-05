package io.seldon.clustermanager.config;

import java.math.BigInteger;
import java.security.SecureRandom;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.oauth2.config.annotation.configurers.ClientDetailsServiceConfigurer;
import org.springframework.security.oauth2.config.annotation.web.configuration.AuthorizationServerConfigurerAdapter;
import org.springframework.security.oauth2.config.annotation.web.configuration.EnableAuthorizationServer;
import org.springframework.security.oauth2.config.annotation.web.configurers.AuthorizationServerEndpointsConfigurer;
import org.springframework.security.oauth2.provider.token.TokenStore;
import org.springframework.security.oauth2.provider.token.store.redis.RedisTokenStore;

@Configuration
@EnableAuthorizationServer
class AuthorizationServerConfiguration extends AuthorizationServerConfigurerAdapter {

    private final static Logger logger = LoggerFactory.getLogger(AuthorizationServerConfiguration.class);
    private final static String SELDON_CLUSTER_MANAGER_CLIENT_SECRET_KEY = "SELDON_CLUSTER_MANAGER_CLIENT_SECRET";
    private final static int ACCESS_TOKEN_VALIDITY_SECONDS = 43200;

    @Autowired
    private AuthenticationManager authenticationManager;

    @Autowired
    private RedisConnectionFactory redisConnectionFactory;

    @Override
    public void configure(AuthorizationServerEndpointsConfigurer endpoints) throws Exception {
        endpoints.tokenStore(tokenStore()).authenticationManager(authenticationManager);
    }

    @Bean
    public TokenStore tokenStore() {
        // return new InMemoryTokenStore();
        return new RedisTokenStore(redisConnectionFactory);
    }

    @Override
    public void configure(ClientDetailsServiceConfigurer clients) throws Exception {

        final String client_id = "client";
        String client_secret = null;
        { // setup the client_secret using the env vars
            client_secret = System.getenv().get(SELDON_CLUSTER_MANAGER_CLIENT_SECRET_KEY);
            if (client_secret == null) {
                client_secret = generateRandomString();
                logger.error(String.format("FAILED to find env var [%s]", SELDON_CLUSTER_MANAGER_CLIENT_SECRET_KEY));
                logger.error(String.format("generating client_secret[%s]", client_secret));
            } else {
                logger.info(String.format("using client_secret from env var [%s]", SELDON_CLUSTER_MANAGER_CLIENT_SECRET_KEY));
            }
        }
        logger.info(String.format("setting up auth using client credentials, client_id[%s]", client_id));

        // @formatter:off
        clients
                .inMemory()
                .withClient(client_id)
                .authorizedGrantTypes("client_credentials", "password")
                .authorities("ROLE_CLIENT")
                .scopes("read","write")
                .resourceIds("cluster-manger-api")
                .accessTokenValiditySeconds(ACCESS_TOKEN_VALIDITY_SECONDS)
                .secret(client_secret);
        // @formatter:on
    }

    private static String generateRandomString() {
        SecureRandom random = new SecureRandom();
        return new BigInteger(130, random).toString(32);
    }
}
