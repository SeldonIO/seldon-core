package io.seldon.apife.metrics;

import static java.util.Arrays.asList;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import io.micrometer.core.instrument.Tag;
import io.micrometer.spring.web.WebmvcTagConfigurer;


public class AuthorizedWebmvcTagConfigurer extends WebmvcTagConfigurer {

	@Override
	public Iterable<Tag> httpRequestTags(HttpServletRequest request,
            HttpServletResponse response,
            Throwable ex) {
		
		return asList(method(request), uri(request), exception(ex), status(response),principal(request));
	}
	
	 public Tag principal(HttpServletRequest request) {
		 if (request.getUserPrincipal() != null)
			 return Tag.of("principal", request.getUserPrincipal().getName());
		 else
			 return Tag.of("principal", "None");
	 }
	
}
