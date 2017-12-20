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

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentCacheImpl implements SeldonDeploymentCache {

	Cache<String, SeldonDeployment> cache = CacheBuilder.newBuilder()
		    .maximumSize(1000)
		    .build();	
	
	private final KubeCRDHandler crdHandler;
	
	@Autowired
	public SeldonDeploymentCacheImpl(KubeCRDHandler crdHandler)
	{
		this.crdHandler = crdHandler;
	}
	
	@Override
    public SeldonDeployment get(String name) {
       return cache.getIfPresent(name);
    }
	
	@Override
	public SeldonDeployment getOrLoad(String name) {
		try {
			return cache.get(name, new Callable<SeldonDeployment>() {
			    @Override
			    public SeldonDeployment call() throws ExecutionException {
			      SeldonDeployment mlDep = SeldonDeploymentCacheImpl.this.crdHandler.getSeldonDeployment(name);
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
		cache.put(dep.getMetadata().getName(), dep);
	}

	@Override
	public void remove(String name) {
		cache.invalidate(name);
		
	}

    
}
