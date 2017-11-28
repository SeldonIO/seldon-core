package io.seldon.clustermanager.k8s.client;

import java.io.IOException;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ProtoClient;

public interface K8sClientProvider {
	public ApiClient getClient() throws IOException;
	public ProtoClient getProtoClient() throws IOException;
}
