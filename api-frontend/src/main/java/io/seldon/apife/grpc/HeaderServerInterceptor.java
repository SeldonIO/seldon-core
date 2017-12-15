package io.seldon.apife.grpc;

import java.util.HashMap;
import java.util.Map;

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.oauth2.common.DefaultOAuth2AccessToken;
import org.springframework.security.oauth2.common.OAuth2AccessToken;
import org.springframework.security.oauth2.provider.OAuth2Authentication;

import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;

/**
 * Validate a gRPC request from an oauth token.
 * @author clive
 *
 */
public class HeaderServerInterceptor implements ServerInterceptor {

    protected static Logger logger = LoggerFactory.getLogger(HeaderServerInterceptor.class.getName());

    public static final String OAUTH_TOKEN_HEADER = "oauth_token";
    
    final Metadata.Key<String> authKey = Metadata.Key.of(OAUTH_TOKEN_HEADER,Metadata.ASCII_STRING_MARSHALLER);

    private SeldonGrpcServer server;
    
    public HeaderServerInterceptor(SeldonGrpcServer server) {
      super();
      this.server = server;
  }

  @Override
  public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
        ServerCall<ReqT, RespT> call,
        final Metadata requestHeaders,
        ServerCallHandler<ReqT, RespT> next) {
      String token = requestHeaders.get(authKey);
      String principal = null;
      if (StringUtils.isEmpty(token))
      {
          logger.warn("Failed to find token");
      }
      else
      {
          Map<String,String> tokenParams = new HashMap<>();
          tokenParams.put(OAuth2AccessToken.ACCESS_TOKEN,token);
          OAuth2AccessToken otoken = DefaultOAuth2AccessToken.valueOf(tokenParams);
          OAuth2Authentication auth = server.getTokenStore().readAuthentication(otoken);
          if (auth != null && auth.isAuthenticated())
          {
              logger.debug("Principal:"+auth.getPrincipal());
              principal = auth.getPrincipal().toString();
          }
          else
          {
              logger.warn("Failed to authenticate token "+token);
          }
      }
      return new MessagePrincipalListener<ReqT>(next.startCall(call, requestHeaders),principal,server);
    }
  }
