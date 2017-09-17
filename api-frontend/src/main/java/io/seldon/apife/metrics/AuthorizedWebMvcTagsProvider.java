package io.seldon.apife.metrics;

import static java.util.Arrays.asList;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import io.micrometer.core.instrument.Tag;
import io.micrometer.spring.web.servlet.DefaultWebMvcTagsProvider;
import io.micrometer.spring.web.servlet.WebMvcTags;


public class AuthorizedWebMvcTagsProvider extends DefaultWebMvcTagsProvider {

	@Override
	public Iterable<Tag> httpRequestTags(HttpServletRequest request,
            HttpServletResponse response,
            Throwable ex) {
		
		return asList(WebMvcTags.method(request), WebMvcTags.uri(request), WebMvcTags.exception(ex), WebMvcTags.status(response),principal(request));
	}
	
	 public Tag principal(HttpServletRequest request) {
		 if (request.getUserPrincipal() != null)
			 return Tag.of("principal", request.getUserPrincipal().getName());
		 else
			 return Tag.of("principal", "None");
	 }

	
	
}
