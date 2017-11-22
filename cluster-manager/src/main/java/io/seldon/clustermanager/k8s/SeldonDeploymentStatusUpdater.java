package io.seldon.clustermanager.k8s;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentStatusUpdater {

	private final SeldonDeploymentCache mlCache;
	private final KubeCRDHandler crdHandler;
	
	@Autowired
	public SeldonDeploymentStatusUpdater(SeldonDeploymentCache mlCache, KubeCRDHandler crdHandler) {
		super();
		this.mlCache = mlCache;
		this.crdHandler = crdHandler;
	}
	
	public void updateStatus(String mlDepName,String depName,int replicasReady)
	{
		SeldonDeployment mlDep = mlCache.get(mlDepName);
		if (mlDep != null)
		{
			SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder(mlDep);
			
			crdHandler.updateSeldonDeployment(mlBuilder.build());
			mlCache.remove(mlDepName);
		}
	}
	
}
