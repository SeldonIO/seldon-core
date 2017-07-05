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
	
	public void addClient(String clientId,ClientDetails cd)
	{
		clientDetails.put(clientId, cd);
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
