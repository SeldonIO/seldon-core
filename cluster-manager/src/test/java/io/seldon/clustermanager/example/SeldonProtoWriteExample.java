package io.seldon.clustermanager.example;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.seldon.protos.DeploymentProtos.ClusterDef;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.PredictiveUnitSubType;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class SeldonProtoWriteExample {

    public static void main(String[] args) {
        System.out.println("Starting...");
        DeploymentDef.Builder deploymentDefBuiler = DeploymentDef.newBuilder();

        deploymentDefBuiler.setCluster(ClusterDef.newBuilder());
        deploymentDefBuiler.setId("1");
        deploymentDefBuiler.setName("my deployment");
        deploymentDefBuiler.setUniqueName("my_interesting_project1.my_deployment.1");

        {
            PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();

            predictorDefBuilder.setEnabled(true);
            predictorDefBuilder.setId("0");
            predictorDefBuilder.setName("my_fantastic_predictor");
            predictorDefBuilder.addImagePullSecrets("my-registry-secret1");
            predictorDefBuilder.addImagePullSecrets("my-registry-secret2");

            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren("1").addChildren("2");

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpu("5")
                        .setGpu("0")
                        .setId("2")
                        .setImage("seldonio/model2")
                        .setMemory("20Gi")
                        .setVersion("1.2")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setType(EndpointDef.EndpointType.REST)
                        );
                //@formatter:on

                predictiveUnitDefBuilder.setId("4");

                predictiveUnitDefBuilder.setName("digit_classifier_v0.2");

                //@formatter:off
                predictiveUnitDefBuilder.addParameters(PredictiveUnitDef.ParamDef.newBuilder()
                            .setName("n_layers")
                            .setType(PredictiveUnitDef.ParamType.STRING)
                            .setValue("5"));
                //@formatter:on

                predictiveUnitDefBuilder.setSubtype(PredictiveUnitSubType.SIMPLE_MODEL);
                predictiveUnitDefBuilder.setType(PredictiveUnitType.MODEL);

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
