package io.seldon.clustermanager.example;

import java.util.List;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.k8s.DeploymentUtils;
import io.seldon.clustermanager.k8s.DeploymentUtils.BuildDeploymentResult;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class DeploymentUtilsExample {

    public static void main(String[] args) {

        DeploymentDef exampleDeploymentDef = KubernetesManagerExampleUtils.buildExampleDeploymentDef();

        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("exampleDeploymentDef:");
            String s = ProtoBufUtils.toJson(exampleDeploymentDef, true);
            System.out.println(s);
            System.out.println("-------------------------------------------------------------------------------");
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        List<BuildDeploymentResult> buildDeploymentResults = DeploymentUtils.buildDeployments(exampleDeploymentDef);
        System.out.println(buildDeploymentResults);

    }
}
