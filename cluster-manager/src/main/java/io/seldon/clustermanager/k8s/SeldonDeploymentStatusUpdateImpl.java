package io.seldon.clustermanager.k8s;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.protos.DeploymentProtos.PredictorStatus;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentStatusUpdateImpl implements SeldonDeploymentStatusUpdate {
    protected static Logger logger = LoggerFactory.getLogger(SeldonDeploymentStatusUpdateImpl.class.getName());
	private final SeldonDeploymentCache mlCache;
	private final KubeCRDHandler crdHandler;
	
	@Autowired
	public SeldonDeploymentStatusUpdateImpl(SeldonDeploymentCache mlCache, KubeCRDHandler crdHandler) {
		super();
		this.mlCache = mlCache;
		this.crdHandler = crdHandler;
	}

	private void update(PredictorStatus.Builder b,Integer replicas, Integer replicasAvailable)
	{
	    if (replicas != null)
	        b.setReplicas(replicas);
	    else
	        b.setReplicas(0);
	    if (replicasAvailable != null)
	        b.setReplicasAvailable(replicasAvailable);
	    else
	        b.setReplicasAvailable(0);
	}
	
    @Override
    public void updateStatus(String mlDepName, String depName, Integer replicas, Integer replicasAvailable) {
        if (replicas == null || replicas == 0)
            removeStatus(mlDepName,depName);
        else
        {
            logger.info(String.format("UPDATE %s : %s %d %d",mlDepName,depName,replicas,replicasAvailable));
            SeldonDeployment mlDep = mlCache.get(mlDepName);
            if (mlDep != null)
            {
                SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder(mlDep);
                
               boolean changed = false;
               for (PredictorStatus.Builder b : mlBuilder.getStatusBuilder().getPredictorStatusBuilderList())
               {
                  if (b.getName().equals(depName))
                  {
                      update(b,replicas,replicasAvailable);
                      changed = true;
                      break;
                  }
               }
               if (!changed)
               {
                   PredictorStatus.Builder b = PredictorStatus.newBuilder().setName(depName);
                   update(b,replicas,replicasAvailable);
                   mlBuilder.getStatusBuilder().addPredictorStatus(b);
               }

               mlCache.remove(mlDepName);
               crdHandler.updateSeldonDeployment(mlBuilder.build());
            }
        }
        
    }

    @Override
    public void removeStatus(String mlDepName, String depName) {
        logger.info(String.format("DELETE %s : %s",mlDepName,depName));
        SeldonDeployment mlDep = mlCache.get(mlDepName);
        if (mlDep != null)
        {
            SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder(mlDep);
            int idx = 0;
            for (PredictorStatus.Builder b : mlBuilder.getStatusBuilder().getPredictorStatusBuilderList())
           {
              if (b.getName().equals(depName))
              {
                  mlBuilder.getStatusBuilder().removePredictorStatus(idx);
                  break;
              }
           }
           mlCache.remove(mlDepName);
           crdHandler.updateSeldonDeployment(mlBuilder.build());
        }
    }
	
}
