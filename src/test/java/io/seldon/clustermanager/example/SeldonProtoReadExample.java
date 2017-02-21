package io.seldon.clustermanager.example;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class SeldonProtoReadExample {

    public static void main(String[] args) {

        DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
        DeploymentDef deploymentDef = null;

        { // From Json
            String json = "{\"cluster\":null,\"id\":1,\"name\":\"my deployment\",\"predictor\":{\"enabled\":true,\"id\":0,\"name\":\"my_fantastic_predictor\",\"predictiveUnits\":[{\"children\":[],\"cluster_resources\":{\"cpus\":5,\"dockerRegistry\":{\"id\":1,\"name\":\"Seldon Registry\",\"password\":\"secret\",\"url\":\"http://registry.seldon.io\",\"username\":\"seldon\"},\"gpus\":0,\"id\":2,\"image\":\"seldonio/model2\",\"memoryGb\":20,\"replicas\":1,\"version\":\"1.2\"},\"endpoint\":{\"host\":\"127.0.0.1\",\"port\":5004,\"type\":\"REST\"},\"id\":4,\"name\":\"digit_classifier_v0.2\",\"parameters\":[{\"name\":\"n_layers\",\"type\":\"INT\",\"value\":\"5\"}],\"subtype\":\"simpleModel\",\"type\":\"model\"}],\"root\":0},\"uniqueName\":\"my_interesting_project1.my_deployment.1\"}";
            try {
                JsonFormat.parser().ignoringUnknownFields().merge(json, deploymentDefBuilder);
                deploymentDef = deploymentDefBuilder.build();

                String depName = deploymentDef.getName();
                System.out.println(String.format("Read Deployment[%s]", depName));
            } catch (InvalidProtocolBufferException e) {
                e.printStackTrace();
            }
        }

        { // back to json
            try {

                String jsonGen = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(deploymentDef);
                System.out.println(jsonGen);
            } catch (InvalidProtocolBufferException e) {
                e.printStackTrace();
            }

        }
    }

}
