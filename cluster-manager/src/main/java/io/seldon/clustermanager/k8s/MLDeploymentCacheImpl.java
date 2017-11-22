package io.seldon.clustermanager.k8s;

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;

import io.seldon.protos.DeploymentProtos.MLDeployment;

@Component
public class MLDeploymentCacheImpl implements MLDeploymentCache {

	Cache<String, MLDeployment> cache = CacheBuilder.newBuilder()
		    .maximumSize(1000)
		    .build();	
	
	private final KubeCRDHandler crdHandler;
	
	@Autowired
	public MLDeploymentCacheImpl(KubeCRDHandler crdHandler)
	{
		this.crdHandler = crdHandler;
	}
	
	public MLDeployment get(String name) {
		try {
			return cache.get(name, new Callable<MLDeployment>() {
			    @Override
			    public MLDeployment call() throws ExecutionException {
			      MLDeployment mlDep = MLDeploymentCacheImpl.this.crdHandler.getMlDeployment(name);
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
	
	public void put(MLDeployment dep) {
		cache.put(dep.getMetadata().getName(), dep);
	}

	@Override
	public void remove(String name) {
		cache.invalidate(name);
		
	}
}
