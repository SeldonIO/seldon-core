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
package io.seldon.apife.deployments;

import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

import javax.annotation.PostConstruct;

import org.apache.commons.lang.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.apife.AppProperties;
import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.apife.k8s.KubernetesUtil;
import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class DeploymentStore implements DeploymentsListener {
	protected static Logger logger = LoggerFactory.getLogger(DeploymentStore.class.getName());
	
	//Oauth key to deployment def
	private ConcurrentMap<String, SeldonDeployment> deploymentStore = new ConcurrentHashMap<>();
	
	private final DeploymentsHandler deploymentsHandler;
	private final InMemoryClientDetailsService clientDetailsService;
	private final AppProperties appProperties;
	private final KubernetesUtil k8sUtil = new KubernetesUtil();
	
	@Autowired
	public DeploymentStore(DeploymentsHandler deploymentsHandler,InMemoryClientDetailsService clientDetailsService,AppProperties appProperties)
	{	
		this.deploymentsHandler = deploymentsHandler;
		this.clientDetailsService = clientDetailsService;
		this.appProperties = appProperties;
	}
	 
	@PostConstruct
	private void init() throws Exception{
		logger.info("Initializing...");
		deploymentsHandler.addListener(this);
	}
	 
	public SeldonDeployment getDeployment(String clientId)
	{
		return deploymentStore.get(clientId);
	}
	  
	 @Override
	 public void deploymentAdded(SeldonDeployment mlDep) {
		 final DeploymentSpec deploymentDef = mlDep.getSpec();
		 final String namespace = k8sUtil.getNamespace(mlDep);
		 
		 if (StringUtils.isEmpty(deploymentDef.getOauthKey()))
		 {
			 logger.warn("Deployment with empty oauth key - ignoring {}",deploymentDef.getName());
		 }
		 else
		 {
			 if (appProperties.isSingleNamespace())
			 {
				 deploymentStore.put(deploymentDef.getOauthKey(), mlDep);
				 clientDetailsService.addClient(deploymentDef.getOauthKey(), deploymentDef.getOauthSecret());
			 }
			 
			 // Always add namespaced key
			 final String namespacedKey = deploymentDef.getOauthKey() + namespace;
			 deploymentStore.put(namespacedKey, mlDep);
			 clientDetailsService.addClient(namespacedKey, deploymentDef.getOauthSecret());			 

			 logger.info("Succesfully added or updated deployment "+deploymentDef.getName());
		 }
	}
	 
	@Override
	public void deploymentUpdated(SeldonDeployment mlDep) {
		deploymentAdded(mlDep);
	}
	
	@Override
	public void deploymentRemoved(SeldonDeployment mlDep) {
		 final DeploymentSpec deploymentDef = mlDep.getSpec();
		 deploymentStore.remove(deploymentDef.getOauthKey());
		 clientDetailsService.removeClient(deploymentDef.getOauthKey());
		 logger.info("Removed deployment "+deploymentDef.getName());
	}
		
}
