package io.seldon.clustermanager.k8s;

import io.seldon.clustermanager.component.KubernetesManager;

public class KubernetesManagerImpl implements KubernetesManager {

    @Override
    public void init() throws Exception {
        System.out.println("KubernetesManager init");
    }

    @Override
    public void cleanup() throws Exception {
        System.out.println("KubernetesManager cleanup");
    }

}
