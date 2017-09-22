package io.seldon.apife.metrics;

import static java.util.Arrays.asList;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.micrometer.core.instrument.Tag;
import io.micrometer.spring.web.servlet.DefaultWebMvcTagsProvider;
import io.micrometer.spring.web.servlet.WebMvcTags;
import io.seldon.apife.deployments.DeploymentStore;

@Component
public class AuthorizedWebMvcTagsProvider extends DefaultWebMvcTagsProvider {

	private final static String PROJECT_ANNOTATION_KEY = "project_name";
	
	@Autowired
	DeploymentStore deploymentStore;
	
	@Override
	public Iterable<Tag> httpRequestTags(HttpServletRequest request,
            HttpServletResponse response,
            Throwable ex) {
		
		final String PrincipalName = request.getUserPrincipal().getName();
		return asList(WebMvcTags.method(request), WebMvcTags.uri(request), WebMvcTags.exception(ex), WebMvcTags.status(response),
					principal(PrincipalName),
					version(PrincipalName),
					predictorName(PrincipalName),
					projectName(PrincipalName));
	}
	
	 public Tag principal(String principalName) {
		 return Tag.of("principal", principalName);
	 }

	 public Tag projectName(String principalName)
	 {
		 return Tag.of("project_name",deploymentStore.getDeployment(principalName).getAnnotationsOrDefault(PROJECT_ANNOTATION_KEY, "unknown"));
	 }
	 
	 public Tag version(String principalName)
	 {
		 return Tag.of("predictor_version",deploymentStore.getDeployment(principalName).getPredictor().getVersion());
	 }

	 public Tag predictorName(String principalName)
	 {
		 return Tag.of("predictor_name",deploymentStore.getDeployment(principalName).getPredictor().getName());
	 }

	

	
}
