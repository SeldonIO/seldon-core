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

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;

import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;

import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentCacheImpl implements SeldonDeploymentCache {

	Cache<CacheKey, SeldonDeployment> cache = CacheBuilder.newBuilder()
		    .maximumSize(1000)
		    .build();	
	
	private final KubeCRDHandler crdHandler;
	
	private static class CacheKey {
		
		private String name;
		private String namespace;
		public CacheKey(String name, String namespace) {
			super();
			this.name = name;
			this.namespace = namespace;
		}
		@Override
		public int hashCode() {
			final int prime = 31;
			int result = 1;
			result = prime * result + ((name == null) ? 0 : name.hashCode());
			result = prime * result + ((namespace == null) ? 0 : namespace.hashCode());
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
			return true;
		}
		
		
	}
	
	@Autowired
	public SeldonDeploymentCacheImpl(ClusterManagerProperites clusterManagerProperites,KubeCRDHandler crdHandler)
	{
		this.crdHandler = crdHandler;
	}
	
	@Override
    public SeldonDeployment get(SeldonDeployment dep) {
       return cache.getIfPresent(new CacheKey(dep.getMetadata().getName(), SeldonDeploymentUtils.getNamespace(dep)));
    }
	
	//@Override
	public SeldonDeployment getOrLoad(String name,String namespace) {
		try {
			return cache.get(new CacheKey(name, namespace), new Callable<SeldonDeployment>() {
			    @Override
			    public SeldonDeployment call() throws ExecutionException {
			      SeldonDeployment mlDep = SeldonDeploymentCacheImpl.this.crdHandler.getSeldonDeployment(name,namespace);
			      if (mlDep == null)
			    	  throw new ExecutionException(null);
			      else
			    	  return mlDep;
			    }
			  });
		} catch (ExecutionException e) {
			return null;
		}
	}
	
	@Override
	public void put(SeldonDeployment dep) {
		cache.put(new CacheKey(dep.getMetadata().getName(), SeldonDeploymentUtils.getNamespace(dep)), dep);
	}

	@Override
	public void remove(SeldonDeployment dep) {
		cache.invalidate(new CacheKey(dep.getMetadata().getName(), SeldonDeploymentUtils.getNamespace(dep)));
		
	}

    
}
