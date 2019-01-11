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
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class AuthorizedWebMvcTagsProvider extends DefaultWebMvcTagsProvider {

	private final static String DEPLOYMENT_NAME_METRIC = "deployment_name";
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
				deploymentName(principalName)
                );
	}
	
	 public Tag principal(String principalName) {
		 if (principalName == null)
			 return Tag.of(PRINCIPAL_METRIC, "None");
		 else
			 return Tag.of(PRINCIPAL_METRIC, principalName);
	 }

	 public Tag deploymentName(String principalName)
	 {
		 SeldonDeployment mlDep = deploymentStore.getDeployment(principalName);
		 if (principalName == null || mlDep == null || !StringUtils.hasText(mlDep.getSpec().getName()))
			 return Tag.of(DEPLOYMENT_NAME_METRIC, "None");
		 else
			 return Tag.of(DEPLOYMENT_NAME_METRIC,mlDep.getSpec().getName());
	 }

	

	
}
