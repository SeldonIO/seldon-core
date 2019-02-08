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
package io.seldon.apife.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.apife.k8s.KubernetesUtil;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;


@Service
public class PredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(PredictionService.class.getName());
		
	@Autowired
	DeploymentStore deploymentStore;
	
	@Autowired
	InternalPredictionService internalPredictionService;
	
	private final KubernetesUtil k8sUtil = new KubernetesUtil();

	public String predict(String request,String clientId)  {

		SeldonDeployment deployment = deploymentStore.getDeployment(clientId);
		if (deployment != null)
		{
			final String endpoint = k8sUtil.getSeldonId(deployment) + "." + k8sUtil.getNamespace(deployment);
			return internalPredictionService.getPrediction(request, endpoint);
		}
		else
			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_NO_RUNNING_DEPLOYMENT,"no deployment with id "+clientId);

	}
	
	public void sendFeedback(String feedback, String deploymentId){
		SeldonDeployment deployment = deploymentStore.getDeployment(deploymentId);
		if (deployment != null)
		{
			final String endpoint = k8sUtil.getSeldonId(deployment) + "." + k8sUtil.getNamespace(deployment);
			internalPredictionService.sendFeedback(feedback, endpoint);
		}
		else
			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_NO_RUNNING_DEPLOYMENT,"no deployment with id "+deploymentId);

	}
}
