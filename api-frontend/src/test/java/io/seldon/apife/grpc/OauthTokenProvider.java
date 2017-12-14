package io.seldon.apife.grpc;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.util.List;

import org.apache.http.HttpResponse;
import org.apache.http.NameValuePair;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.ClientProtocolException;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.HttpClient;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.protocol.HttpClientContext;
import org.apache.http.client.utils.URIBuilder;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.protocol.HttpContext;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.seldon.apife.service.InternalPredictionService;

public class OauthTokenProvider {
    private static Logger logger = LoggerFactory.getLogger(OauthTokenProvider.class.getName());

    
    public String getToken() throws URISyntaxException, ClientProtocolException, IOException
    {
        
        CredentialsProvider provider = new BasicCredentialsProvider();
        UsernamePasswordCredentials credentials
         = new UsernamePasswordCredentials("key", "secret");
        provider.setCredentials(AuthScope.ANY, credentials);
          
        HttpClient client = HttpClientBuilder.create()
          .setDefaultCredentialsProvider(provider)
          .build();
         
        URIBuilder builder = new URIBuilder().setScheme("http")
                .setHost("0.0.0.0")
                .setPort(8080)
                .setPath("/oauth/token");

        URI uri = builder.build();
   
       
        //StringEntity requestEntity = new StringEntity("grant_type=client_credentials",ContentType.APPLICATION_JSON);
    
        HttpContext context = HttpClientContext.create();
        HttpPost httpPost = new HttpPost(uri);

        List<NameValuePair> params = new ArrayList<NameValuePair>();
        params.add(new BasicNameValuePair("grant_type", "client_credentials"));
        httpPost.setEntity(new UrlEncodedFormEntity(params));

        HttpResponse resp = client.execute(httpPost, context);
        
        
        String tokenResp = EntityUtils.toString(resp.getEntity());   
        
        ObjectMapper mapper = new ObjectMapper();
        JsonNode actualObj = mapper.readTree(tokenResp);
        String token = actualObj.get("access_token").asText();
        
       return token;
    }
    
    public static void main(String[] args) throws ClientProtocolException, URISyntaxException, IOException {
        OauthTokenProvider p = new OauthTokenProvider();
        String token = p.getToken();
        logger.info(token);
    }
    
}
