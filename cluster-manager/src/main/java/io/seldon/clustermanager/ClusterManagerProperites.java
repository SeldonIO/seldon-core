package io.seldon.clustermanager;

import org.apache.commons.lang3.builder.ReflectionToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;

public class ClusterManagerProperites {

    private int engineContainerPort;
    private int engineGrpcContainerPort;
    private String engineContainerImageAndVersion;
    private int puContainerPortBase;
    private boolean istioEnabled;

    public int getEngineContainerPort() {
        return engineContainerPort;
    }

    public void setEngineContainerPort(int engineContainerPort) {
        this.engineContainerPort = engineContainerPort;
    }

    public int getEngineGrpcContainerPort() {
        return engineGrpcContainerPort;
    }

    public void setEngineGrpcContainerPort(int engineGrpcContainerPort) {
        this.engineGrpcContainerPort = engineGrpcContainerPort;
    }

    public String getEngineContainerImageAndVersion() {
        return engineContainerImageAndVersion;
    }

    public void setEngineContainerImageAndVersion(String engineContainerImageAndVersion) {
        this.engineContainerImageAndVersion = engineContainerImageAndVersion;
    }

    public int getPuContainerPortBase() {
        return puContainerPortBase;
    }

    public void setPuContainerPortBase(int puContainerPortBase) {
        this.puContainerPortBase = puContainerPortBase;
    }
    
    public boolean isIstioEnabled() {
		return istioEnabled;
	}

	public void setIstioEnabled(boolean istioEnabled) {
		this.istioEnabled = istioEnabled;
	}

	@Override
    public String toString() {
        return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
    }
}
