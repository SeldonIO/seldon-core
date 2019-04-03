package io.seldon.clustermanager.k8s.client;

import org.springframework.stereotype.Component;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.apis.AutoscalingV2beta1Api;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;

@Component
public class K8sDefaultApiProvider implements K8sApiProvider {

	@Override
	public CustomObjectsApi getCustomObjectsApi(ApiClient client) {
		return new CustomObjectsApi(client);
	}

	@Override
	public ExtensionsV1beta1Api getExtensionsV1beta1Api(ApiClient client) {
		return new ExtensionsV1beta1Api(client);
	}

	@Override
	public AutoscalingV2beta1Api getAutoScalingApi(ApiClient client) {
		return new AutoscalingV2beta1Api(client);
	}
	
	@Override
	public CoreV1Api getCoreV1Api(ApiClient client) {
		return new CoreV1Api(client);
	}

}
