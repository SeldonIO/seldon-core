package io.seldon.clustermanager.k8s;

import io.fabric8.kubernetes.api.model.OwnerReference;

public class CustomResourceDetails {
	
	private int resourceVersion;
	private OwnerReference oref;
	
	public CustomResourceDetails(int resourceVersion, OwnerReference oref) {
		super();
		this.resourceVersion = resourceVersion;
		this.oref = oref;
	}
	public int getResourceVersion() {
		return resourceVersion;
	}
	public OwnerReference getOref() {
		return oref;
	}
	
	
}
