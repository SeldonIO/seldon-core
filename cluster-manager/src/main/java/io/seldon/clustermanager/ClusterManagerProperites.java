package io.seldon.clustermanager;

import org.apache.commons.lang3.builder.ReflectionToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;

public class ClusterManagerProperites {

    private int engineContainerPort;
    private String engineContainerImageAndVersion;
    private int puContainerPortBase;

    public int getEngineContainerPort() {
        return engineContainerPort;
    }

    public void setEngineContainerPort(int engineContainerPort) {
        this.engineContainerPort = engineContainerPort;
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

    @Override
    public String toString() {
        return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
    }
}
