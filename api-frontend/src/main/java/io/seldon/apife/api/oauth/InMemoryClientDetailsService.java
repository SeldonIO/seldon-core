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
package io.seldon.apife.api.oauth;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

import org.springframework.security.oauth2.provider.ClientDetails;
import org.springframework.security.oauth2.provider.ClientDetailsService;
import org.springframework.security.oauth2.provider.ClientRegistrationException;
import org.springframework.security.oauth2.provider.NoSuchClientException;
import org.springframework.stereotype.Component;

@Component
public class InMemoryClientDetailsService implements ClientDetailsService {

	private Map<String, ClientDetails> clientDetails = new ConcurrentHashMap<String, ClientDetails>();
	private final static int ACCESS_TOKEN_VALIDITY_SECONDS = 43200;
	public void addClient(String clientId,String secret)
	{
		 ClientBuilder cb = new ClientBuilder(clientId);
		 cb.authorizedGrantTypes("client_credentials", "password")
		 	.authorities("ROLE_CLIENT")
	        .scopes("read","write")
	        .resourceIds("prediction-client")
	        .accessTokenValiditySeconds(ACCESS_TOKEN_VALIDITY_SECONDS)
	        .secret(secret);	
		 
		 this.addClient(clientId, cb.build());
	}	
	
	public void addClient(String clientId,ClientDetails cd)
	{
		clientDetails.put(clientId, cd);
	}
	
	public void removeClient(String clientId)
	{
		clientDetails.remove(clientId);
	}
	
	@Override
	public ClientDetails loadClientByClientId(String clientId) throws ClientRegistrationException {
	    ClientDetails details = clientDetails.get(clientId);
	    if (details == null) {
	      throw new NoSuchClientException("No client with requested id: " + clientId);
	    }
	    return details;
	  }


}
