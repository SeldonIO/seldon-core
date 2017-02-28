package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.List;

import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientException;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.CMStatusDef;
import io.seldon.protos.DeploymentProtos.StringListDef;

public class KubernetesManagerImpl implements KubernetesManager {

    private KubernetesClient kubernetesClient = null;

    @Override
    public void init() throws Exception {
        System.out.println("KubernetesManager: init");

        String master = "http://localhost:8001/";
        Config config = new ConfigBuilder().withMasterUrl(master).build();

        try {
            kubernetesClient = new DefaultKubernetesClient(config);
            getNamespaceList(); // simple check to see if client works
            System.out.println("KubernetesManager: sucessfully passed namespace check test");
        } catch (KubernetesClientException e) {
            throw new Exception(e);
        }
    }

    @Override
    public void cleanup() throws Exception {
        System.out.println("KubernetesManager: cleanup");
        if (kubernetesClient != null) {
            kubernetesClient.close();
        }
    }

    private List<String> getNamespaceList() {
        List<String> namespace_list = new ArrayList<>();
        NamespaceList namespaceList = kubernetesClient.namespaces().list();
        for (Namespace ns : namespaceList.getItems()) {
            namespace_list.add(ns.getMetadata().getName());
        }

        return namespace_list;
    }

    @Override
    public CMResultDef getNamespaces() {
        CMResultDef cmResultDef = null;
        try {
            List<String> namespace_list = new ArrayList<>();
            NamespaceList namespaceList = kubernetesClient.namespaces().list();
            for (Namespace ns : namespaceList.getItems()) {
                namespace_list.add(ns.getMetadata().getName());
            }
            
            //@formatter:off
            StringListDef.Builder stringListDefBuilder = StringListDef.newBuilder();
            for (String item: namespace_list) {
                stringListDefBuilder.addItems(item);
            }
            StringListDef stringListDef = stringListDefBuilder.build();
                    
            cmResultDef = CMResultDef.newBuilder()
                    .setCmstatus(CMStatusDef.newBuilder()
                            .setCode(200)
                            .setStatus(CMStatusDef.Status.SUCCESS))
                    .setStringList(stringListDef)
                    .build();
            //@formatter:on

        } catch (Throwable e) {
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            //@formatter:off
            cmResultDef = CMResultDef.newBuilder()
                    .setCmstatus(CMStatusDef.newBuilder()
                            .setCode(500)
                            .setStatus(CMStatusDef.Status.FAILURE)
                            .setInfo(info))
                    .clearStringList()
                    .build();
            //@formatter:on
        }

        return cmResultDef;
    }
}
