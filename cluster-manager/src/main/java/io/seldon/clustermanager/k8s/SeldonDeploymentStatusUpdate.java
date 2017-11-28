package io.seldon.clustermanager.k8s;

public interface SeldonDeploymentStatusUpdate {
    public void updateStatus(String mlDepName,String depName,Integer replicas,Integer replicasReady);
    public void removeStatus(String mlDepName,String depName);
}
