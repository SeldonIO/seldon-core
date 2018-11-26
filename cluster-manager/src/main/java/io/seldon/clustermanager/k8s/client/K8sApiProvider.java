package io.seldon.clustermanager.k8s.client;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;

public interface K8sApiProvider {

	public CustomObjectsApi getCustomObjectsApi(ApiClient client);
	public ExtensionsV1beta1Api getExtensionsV1beta1Api(ApiClient client);
}
