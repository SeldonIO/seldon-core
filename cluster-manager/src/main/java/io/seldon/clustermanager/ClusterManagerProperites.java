package io.seldon.clustermanager;

import org.apache.commons.lang3.builder.ReflectionToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;

public class ClusterManagerProperites {

    private int engineContainerPort;

    public int getEngineContainerPort() {
        return engineContainerPort;
    }

    public void setEngineContainerPort(int engineContainerPort) {
        this.engineContainerPort = engineContainerPort;
    }

    @Override
    public String toString() {
        return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
    }

}
