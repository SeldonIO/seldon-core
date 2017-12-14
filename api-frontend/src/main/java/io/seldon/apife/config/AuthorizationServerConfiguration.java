package io.seldon.apife.config;

import java.math.BigInteger;
import java.security.SecureRandom;

import javax.sql.DataSource;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.jdbc.datasource.DriverManagerDataSource;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.oauth2.config.annotation.configurers.ClientDetailsServiceConfigurer;
import org.springframework.security.oauth2.config.annotation.web.configuration.AuthorizationServerConfigurerAdapter;
import org.springframework.security.oauth2.config.annotation.web.configuration.EnableAuthorizationServer;
import org.springframework.security.oauth2.config.annotation.web.configurers.AuthorizationServerEndpointsConfigurer;
import org.springframework.security.oauth2.provider.token.TokenStore;
import org.springframework.security.oauth2.provider.token.store.redis.RedisTokenStore;

import io.seldon.apife.api.oauth.ClientBuilder;
import io.seldon.apife.api.oauth.InMemoryClientDetailsService;

@Configuration
@EnableAuthorizationServer
class AuthorizationServerConfiguration extends AuthorizationServerConfigurerAdapter {

    private final static Logger logger = LoggerFactory.getLogger(AuthorizationServerConfiguration.class);
    private final static String TEST_CLIENT_KEY = "TEST_CLIENT_KEY";
    private final static String TEST_CLIENT_SECRET = "TEST_CLIENT_SECRET";
   

    @Autowired
    private AuthenticationManager authenticationManager;

    @Autowired
    private RedisConnectionFactory redisConnectionFactory;

    @Autowired
    private InMemoryClientDetailsService clientDetailsService;
    
    @Override
    public void configure(AuthorizationServerEndpointsConfigurer endpoints) throws Exception {
        endpoints.tokenStore(tokenStore()).authenticationManager(authenticationManager);
    }

    @Bean
    public TokenStore tokenStore() {
        // return new InMemoryTokenStore();
        return new RedisTokenStore(redisConnectionFactory);
    }
    
    public AuthenticationManager getAuthenticationManager()
    {
        return authenticationManager;
    }
    
    
    /*
    @Bean
    public DataSource dataSource() {
        final DriverManagerDataSource dataSource = new DriverManagerDataSource();
        dataSource.setDriverClassName("com.mysql.jdbc.Driver");
        dataSource.setUrl("jdbc:mysql://localhost/oauth_client");
        dataSource.setUsername("user1");
        dataSource.setPassword("mypass");
        return dataSource;
	}
	*/

    @Override
    public void configure(ClientDetailsServiceConfigurer clients) throws Exception {
        
        clients.withClientDetails(clientDetailsService);
        /*
         		.jdbc(dataSource());
         */
        /*
                .inMemory()
                .withClient(client_id)
                .authorizedGrantTypes("client_credentials", "password")
                .authorities("ROLE_CLIENT")
                .scopes("read","write")
                .resourceIds("cluster-manger-api")
                .accessTokenValiditySeconds(ACCESS_TOKEN_VALIDITY_SECONDS)
                .secret(client_secret);
                */
        // @formatter:on
        
        String client_key = System.getenv().get(TEST_CLIENT_KEY);
        if (client_key != null)
        {
            String client_secret = System.getenv().get(TEST_CLIENT_SECRET);
            clientDetailsService.addClient(client_key,client_secret);
        }
    }

    private static String generateRandomString() {
        SecureRandom random = new SecureRandom();
        return new BigInteger(130, random).toString(32);
    }
}
