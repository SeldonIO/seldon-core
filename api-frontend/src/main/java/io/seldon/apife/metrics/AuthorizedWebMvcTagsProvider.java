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
	private final static String DEPLOYMENT_NAME_METRIC = "deployment_name";
	private final static String DEPLOYMENT_VERSION_METRIC = "deployment_version";
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
                projectName(principalName),
				deploymentName(principalName),
                deploymentVersion(principalName)
                );

		
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
	 
	 public Tag deploymentVersion(String principalName)
	 {
		 if (principalName == null || !StringUtils.hasText(deploymentStore.getDeployment(principalName).getAnnotationsOrDefault(DEPLOYMENT_VERSION_METRIC, "")))
			 return Tag.of(DEPLOYMENT_VERSION_METRIC, "None");
		 else
			 return Tag.of(DEPLOYMENT_VERSION_METRIC,deploymentStore.getDeployment(principalName).getAnnotationsOrDefault(DEPLOYMENT_VERSION_METRIC, ""));
	 }

	 public Tag deploymentName(String principalName)
	 {
		 if (principalName == null || !StringUtils.hasText(deploymentStore.getDeployment(principalName).getName()))
			 return Tag.of(DEPLOYMENT_NAME_METRIC, "None");
		 else
			 return Tag.of(DEPLOYMENT_NAME_METRIC,deploymentStore.getDeployment(principalName).getName());
	 }

	

	
}
