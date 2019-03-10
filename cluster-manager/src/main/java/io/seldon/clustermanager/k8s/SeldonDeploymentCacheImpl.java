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
package io.seldon.clustermanager.k8s;

import org.springframework.stereotype.Component;

import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

/**
 * Reference implementation for Seldon Deployment cache.
 * @author clive
 *
 */
@Component
public class SeldonDeploymentCacheImpl implements SeldonDeploymentCache {

	Cache<CacheKey, SeldonDeployment> cache = CacheBuilder.newBuilder()
		    .maximumSize(1000)
		    .build();	
	
	/**
	 * A key derived to ensure unique references basedon name, the api version and namespace of the Seldon Deployment
	 * @author clive
	 *
	 */
	private static class CacheKey {
		
		private String name;
		private String version;
		private String namespace;
		public CacheKey(String name, String version, String namespace) {
			super();
			this.name = name;
			this.version = version;
			this.namespace = namespace;
		}
		@Override
		public int hashCode() {
			final int prime = 31;
			int result = 1;
			result = prime * result + ((name == null) ? 0 : name.hashCode());
			result = prime * result + ((namespace == null) ? 0 : namespace.hashCode());
			result = prime * result + ((version == null) ? 0 : version.hashCode());
			return result;
		}
		@Override
		public boolean equals(Object obj) {
			if (this == obj)
				return true;
			if (obj == null)
				return false;
			if (getClass() != obj.getClass())
				return false;
			CacheKey other = (CacheKey) obj;
			if (name == null) {
				if (other.name != null)
					return false;
			} else if (!name.equals(other.name))
				return false;
			if (namespace == null) {
				if (other.namespace != null)
					return false;
			} else if (!namespace.equals(other.namespace))
				return false;
			if (version == null) {
				if (other.version != null)
					return false;
			} else if (!version.equals(other.version))
				return false;
			return true;
		}
	}
		
	@Override
    public SeldonDeployment get(SeldonDeployment dep) {
       return cache.getIfPresent(new CacheKey(dep.getMetadata().getName(), dep.getApiVersion(), SeldonDeploymentUtils.getNamespace(dep)));
    }
	
	@Override
	public void put(SeldonDeployment dep) {
		cache.put(new CacheKey(dep.getMetadata().getName(), dep.getApiVersion(), SeldonDeploymentUtils.getNamespace(dep)), dep);
	}

	@Override
	public void remove(SeldonDeployment dep) {
		cache.invalidate(new CacheKey(dep.getMetadata().getName(), dep.getApiVersion(), SeldonDeploymentUtils.getNamespace(dep)));
		
	}

    
}
