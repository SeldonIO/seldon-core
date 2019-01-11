/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.clustermanager;

import org.apache.commons.lang3.builder.ReflectionToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;

public class ClusterManagerProperites {

    private int engineContainerPort;
    private int engineGrpcContainerPort;
    private String engineContainerImageAndVersion;
    private String engineContainerImagePullPolicy = "IfNotPresent";
    private String engineContainerServiceAccountName = "default";
    private int puContainerPortBase;
    private String namespace;
    private boolean singleNamespace = true;
    private int engineUser = 8889;

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

    public String getEngineContainerImagePullPolicy() {
        return engineContainerImagePullPolicy;
    }

    public void setEngineContainerImagePullPolicy(String engineContainerImagePullPolicy) {
        this.engineContainerImagePullPolicy = engineContainerImagePullPolicy;
    }
    
    public String getEngineContainerServiceAccountName() {
		return engineContainerServiceAccountName;
	}

	public void setEngineContainerServiceAccountName(String engineContainerServiceAccountName) {
		this.engineContainerServiceAccountName = engineContainerServiceAccountName;
	}

	public int getPuContainerPortBase() {
        return puContainerPortBase;
    }

    public void setPuContainerPortBase(int puContainerPortBase) {
        this.puContainerPortBase = puContainerPortBase;
    }
    
	public String getNamespace() {
		return namespace;
	}

	public void setNamespace(String namespace) {
		this.namespace = namespace;
	}
	
	public boolean isSingleNamespace() {
		return singleNamespace;
	}

	public void setSingleNamespace(boolean singleNamespace) {
		this.singleNamespace = singleNamespace;
	}

	public int getEngineUser() {
		return engineUser;
	}

	public void setEngineUser(int engineUser) {
		this.engineUser = engineUser;
	}

	@Override
    public String toString() {
        return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
    }
}
