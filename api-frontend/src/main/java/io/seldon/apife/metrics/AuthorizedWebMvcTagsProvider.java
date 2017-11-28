package io.seldon.apife.metrics;

import static java.util.Arrays.asList;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import io.micrometer.core.instrument.Tag;
import io.micrometer.spring.web.servlet.DefaultWebMvcTagsProvider;
import io.micrometer.spring.web.servlet.WebMvcTags;
import io.seldon.apife.deployments.DeploymentStore;

@Component
public class AuthorizedWebMvcTagsProvider extends DefaultWebMvcTagsProvider {

	private final static String PROJECT_ANNOTATION_KEY = "project_name";
	private final static String PREDICTOR_NAME_METRIC = "predictor_name";
	private final static String PREDICTOR_VERSION_METRIC = "predictor_version";
	private final static String PRINCIPAL_METRIC = "principal";
	
	@Autowired
	DeploymentStore deploymentStore;
	
	@Override
	public Iterable<Tag> httpRequestTags(HttpServletRequest request,
            HttpServletResponse response,
            Throwable ex) {
		
		String principalName = null;
		if (request.getUserPrincipal() != null )
			principalName = request.getUserPrincipal().getName();

		return asList(WebMvcTags.method(request), WebMvcTags.uri(request), WebMvcTags.exception(ex), WebMvcTags.status(response),
				principal(principalName),
				version(principalName),
				predictorName(principalName),
				projectName(principalName));
		
	}
	
	 public Tag principal(String principalName) {
		 if (principalName == null)
			 return Tag.of(PRINCIPAL_METRIC, "None");
		 else
			 return Tag.of(PRINCIPAL_METRIC, principalName);
	 }

	 public Tag projectName(String principalName)
	 {
		 if (principalName != null)
			 return Tag.of(PROJECT_ANNOTATION_KEY,deploymentStore.getDeployment(principalName).getAnnotationsOrDefault(PROJECT_ANNOTATION_KEY, "unknown"));
		 else
			 return Tag.of(PROJECT_ANNOTATION_KEY, "None");
	 }
	 
	 public Tag version(String principalName)
	 {
		 if (principalName == null || !StringUtils.hasText(deploymentStore.getDeployment(principalName).getAnnotationsOrDefault("version", "")))
			 return Tag.of(PREDICTOR_VERSION_METRIC, "None");
		 else
			 return Tag.of(PREDICTOR_VERSION_METRIC,deploymentStore.getDeployment(principalName).getAnnotationsOrDefault("version", ""));
	 }

	 public Tag predictorName(String principalName)
	 {
		 if (principalName == null || !StringUtils.hasText(deploymentStore.getDeployment(principalName).getName()))
			 return Tag.of(PREDICTOR_NAME_METRIC, "None");
		 else
			 return Tag.of(PREDICTOR_NAME_METRIC,deploymentStore.getDeployment(principalName).getName());
	 }

	

	
}
