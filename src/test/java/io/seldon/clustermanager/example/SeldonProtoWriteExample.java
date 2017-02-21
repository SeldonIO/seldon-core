package io.seldon.clustermanager.example;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.seldon.protos.DeploymentProtos.ClusterDef;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class SeldonProtoWriteExample {

    public static void main(String[] args) {
        System.out.println("Starting...");
        DeploymentDef.Builder deploymentDefBuiler = DeploymentDef.newBuilder();

        deploymentDefBuiler.setCluster(ClusterDef.newBuilder());
        deploymentDefBuiler.setId(1);
        deploymentDefBuiler.setName("my deployment");
        deploymentDefBuiler.setUniqueName("my_interesting_project1.my_deployment.1");

        {
            PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();

            predictorDefBuilder.setEnabled(true);
            predictorDefBuilder.setId(0);
            predictorDefBuilder.setName("my_fantastic_predictor");

            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren(1).addChildren(2);

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpus(5)
                        .setDockerRegistry(ClusterResourcesDef.DockerRegistryDef.newBuilder()
                                .setId(1).setName("Seldon Registry")
                                .setPassword("secret")
                                .setUrl("http://registry.seldon.io")
                                .setUsername("seldon")
                                .build())
                        .setGpus(0)
                        .setId(2)
                        .setImage("seldonio/model2")
                        .setMemoryGb(20)
                        .setReplicas(1)
                        .setVersion("1.2")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setHost("127.0.0.1")
                        .setPort(5004)
                        .setType(EndpointDef.EndpointType.REST)
                        );
                //@formatter:on

                predictiveUnitDefBuilder.setId(4);

                predictiveUnitDefBuilder.setName("digit_classifier_v0.2");

                //@formatter:off
                predictiveUnitDefBuilder.addParameters(PredictiveUnitDef.ParamDef.newBuilder()
                            .setName("n_layers")
                            .setType(PredictiveUnitDef.ParamType.STRING)
                            .setValue("5"));
                //@formatter:on

                predictiveUnitDefBuilder.setSubtype("simpleModel");

                predictiveUnitDefBuilder.setType("model");

                PredictiveUnitDef predictiveUnitDef = predictiveUnitDefBuilder.build();
                predictorDefBuilder.addPredictiveUnits(predictiveUnitDef);
            }

            PredictorDef predictorDef = predictorDefBuilder.build();

            deploymentDefBuiler.setPredictor(predictorDef);

        }

        DeploymentDef deploymentDef = deploymentDefBuiler.build();

        try {

            /// JsonFormat.TypeRegistry registry =
            /// JsonFormat.TypeRegistry.newBuilder().add(PredictiveUnitDef.StringParamDef.getDescriptor()).build();
            // String json =
            /// JsonFormat.printer().includingDefaultValueFields().usingTypeRegistry(registry).print(deploymentDef);

            String json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(deploymentDef);
            System.out.println(json);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        System.out.println("Done");
    }

}
