package io.seldon.clustermanager.k8s;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.protos.DeploymentProtos.MLDeployment;
import io.seldon.protos.DeploymentProtos.MLDeploymentStatus;

@Component
public class MLDeploymentStatusUpdater {

	private final MLDeploymentCache mlCache;
	private final KubeCRDHandler crdHandler;
	
	@Autowired
	public MLDeploymentStatusUpdater(MLDeploymentCache mlCache, KubeCRDHandler crdHandler) {
		super();
		this.mlCache = mlCache;
		this.crdHandler = crdHandler;
	}
	
	public void updateStatus(String mlDepName,String depName,int replicasReady)
	{
		MLDeployment mlDep = mlCache.get(mlDepName);
		if (mlDep != null)
		{
			boolean isCanary = false;
			if (depName.endsWith("c"))
				isCanary = true;
			MLDeployment.Builder mlBuilder = MLDeployment.newBuilder(mlDep);
			if (!isCanary)
			{
				if (replicasReady == mlDep.getStatus().getPredictorReplicasReady())
					return;
				mlBuilder.setStatus(MLDeploymentStatus.newBuilder(mlDep.getStatus()).setPredictorReplicasReady(replicasReady).build());
			}
			else
			{	
				if (replicasReady == mlDep.getStatus().getCanaryReplicasReady() || !mlDep.getSpec().hasPredictorCanary())
					return;
				mlBuilder.setStatus(MLDeploymentStatus.newBuilder(mlDep.getStatus()).setCanaryReplicasReady(replicasReady).build());
			}
			crdHandler.updateMLDeployment(mlBuilder.build());
			mlCache.remove(mlDepName);
		}
	}
	
}
