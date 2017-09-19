package io.seldon.apife.metrics;

import static java.util.Arrays.asList;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import org.springframework.beans.factory.annotation.Autowired;

import io.micrometer.core.instrument.Tag;
import io.micrometer.spring.web.servlet.DefaultWebMvcTagsProvider;
import io.micrometer.spring.web.servlet.WebMvcTags;
import io.seldon.apife.deployments.DeploymentStore;


public class AuthorizedWebMvcTagsProvider extends DefaultWebMvcTagsProvider {

	@Autowired
	DeploymentStore deploymentStore;
	
	@Override
	public Iterable<Tag> httpRequestTags(HttpServletRequest request,
            HttpServletResponse response,
            Throwable ex) {
		
		return asList(WebMvcTags.method(request), WebMvcTags.uri(request), WebMvcTags.exception(ex), WebMvcTags.status(response),
					principal(request),
					version(request),
					predictorName(request),
					projectName(request));
	}
	
	 public Tag principal(HttpServletRequest request) {
		 if (request.getUserPrincipal() != null)
			 return Tag.of("principal", request.getUserPrincipal().getName());
		 else
			 return Tag.of("principal", "None");
	 }

	 public Tag version(HttpServletRequest request)
	 {
		 return Tag.of("version",deploymentStore.getDeployment(request.getUserPrincipal().getName()).getPredictor().getVersion());
	 }

	 public Tag predictorName(HttpServletRequest request)
	 {
		 return Tag.of("predictor_name",deploymentStore.getDeployment(request.getUserPrincipal().getName()).getPredictor().getName());
	 }

	 public Tag projectName(HttpServletRequest request)
	 {
		 return Tag.of("project_name",deploymentStore.getDeployment(request.getUserPrincipal().getName()).getProjectName());
	 }

	
}
