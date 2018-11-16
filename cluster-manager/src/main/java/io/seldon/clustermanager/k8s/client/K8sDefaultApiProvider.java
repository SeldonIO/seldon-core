package io.seldon.clustermanager.k8s.client;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;

public class K8sDefaultApiProvider implements K8sApiProvider {

	@Override
	public CustomObjectsApi getCustomObjectsApi(ApiClient client) {
		return new CustomObjectsApi(client);
	}

	@Override
	public ExtensionsV1beta1Api getExtensionsV1beta1Api(ApiClient client) {
		return new ExtensionsV1beta1Api(client);
	}

}
