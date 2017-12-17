package io.seldon.apife.config;

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

import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

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

    @Autowired
    private DeploymentStore deploymentStore;
       
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
    
    @Override
    public void configure(ClientDetailsServiceConfigurer clients) throws Exception {
        
        clients.withClientDetails(clientDetailsService);
 
        String client_key = System.getenv().get(TEST_CLIENT_KEY);
        //Create Fake seldon deployment for testing
        if (client_key != null)
        {
            String client_secret = System.getenv().get(TEST_CLIENT_SECRET);
            clientDetailsService.addClient(client_key,client_secret);
            SeldonDeployment dep = SeldonDeployment.newBuilder()
                    .setApiVersion("v1alpha1")
                    .setKind("SeldonDeplyment")
                    .setSpec(DeploymentSpec.newBuilder()
                            .setName("localhost")
                        .setOauthKey(client_key)
                        .setOauthSecret(client_secret)
                        ).build();   
            deploymentStore.deploymentAdded(dep);
        }
    }
}
