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
package io.seldon.clustermanager.k8s;

import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1ServiceList;
import io.kubernetes.client.models.V2beta1HorizontalPodAutoscalerList;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

/**
 * Interactions to update and get details about Seldon deployments and associated resources
 * @author clive
 *
 */
public interface KubeCRDHandler {

	public void updateRaw(String json, String seldonDeploymentName, String version, String namespace);
	public void updateSeldonDeploymentStatus(SeldonDeployment mlDep);
	public SeldonDeployment getSeldonDeployment(String name, String version, String namespace);	
	public ExtensionsV1beta1DeploymentList getOwnedDeployments(String seldonDeploymentName,String namespace);
	public V1ServiceList getOwnedServices(String seldonDeploymentName,String namespace);
	public V2beta1HorizontalPodAutoscalerList getOwnedHPAs(String seldonDeploymentName,String namespace);
}
