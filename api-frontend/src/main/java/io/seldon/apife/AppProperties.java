package io.seldon.apife;

import org.apache.commons.lang3.builder.ReflectionToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;

public class AppProperties {

    private int engineContainerPort;
    private int engineGrpcContainerPort;
    
    
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

    
    @Override
    public String toString() {
        return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
    }
    
}
