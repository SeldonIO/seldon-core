package io.seldon.clustermanager.k8s.client;

import java.io.IOException;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.util.Config;

public class K8sDefaultClientProvider implements K8sClientProvider {

	@Override
	public ApiClient getClient() throws IOException {
		return Config.defaultClient();
	}

}
