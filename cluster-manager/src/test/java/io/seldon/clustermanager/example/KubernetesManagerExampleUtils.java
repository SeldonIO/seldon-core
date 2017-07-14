package io.seldon.clustermanager.example;

import io.seldon.protos.DeploymentProtos.ClusterDef;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.PredictiveUnitSubType;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class KubernetesManagerExampleUtils {

    public static DeploymentDef buildExampleDeploymentDef() {
        DeploymentDef.Builder deploymentDefBuiler = DeploymentDef.newBuilder();

        deploymentDefBuiler.setCluster(ClusterDef.newBuilder());
        deploymentDefBuiler.setId("1");
        deploymentDefBuiler.setName("my deployment");
        deploymentDefBuiler.setUniqueName("my-interesting-project1.my-deployment.1");

        PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();
        {

            predictorDefBuilder.setEnabled(true);
            predictorDefBuilder.setId("0");
            predictorDefBuilder.setName("my_fantastic_predictor");
            predictorDefBuilder.setReplicas(3);
            predictorDefBuilder.setReplicasReady(0);
            predictorDefBuilder.addImagePullSecrets("my-registry-secret1");
            predictorDefBuilder.addImagePullSecrets("my-registry-secret2");

            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren("1").addChildren("2");

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpu("0.1")
                        .setGpu("0")
                        .setId("2")
                        .setImage("bogusimage")
                        .setMemory("0.5Gi")
                        .setVersion("")
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
                predictiveUnitDefBuilder.addParameters(PredictiveUnitDef.ParamDef.newBuilder()
                            .setName("x_layers")
                            .setType(PredictiveUnitDef.ParamType.STRING)
                            .setValue("6"));
                //@formatter:on

                predictiveUnitDefBuilder.setSubtype(PredictiveUnitSubType.SIMPLE_MODEL);
                predictiveUnitDefBuilder.setType(PredictiveUnitType.MODEL);

                PredictiveUnitDef predictiveUnitDef = predictiveUnitDefBuilder.build();
                predictorDefBuilder.addPredictiveUnits(predictiveUnitDef);
            }
            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren("1").addChildren("2");

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpu("0.1")
                        .setGpu("0")
                        .setId("2")
                        .setImage("gsunner/putest")
                        .setMemory("0.1Gi")
                        .setVersion("")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setType(EndpointDef.EndpointType.REST)
                        );
                //@formatter:on

                predictiveUnitDefBuilder.setId("6");

                predictiveUnitDefBuilder.setName("digit_classifier_v0.2");

                //@formatter:off
                predictiveUnitDefBuilder.addParameters(PredictiveUnitDef.ParamDef.newBuilder()
                            .setName("n_layers")
                            .setType(PredictiveUnitDef.ParamType.STRING)
                            .setValue("5"));
                //@formatter:on

                predictiveUnitDefBuilder.setSubtype(PredictiveUnitSubType.MICROSERVICE);
                predictiveUnitDefBuilder.setType(PredictiveUnitType.MODEL);

                PredictiveUnitDef predictiveUnitDef = predictiveUnitDefBuilder.build();
                predictorDefBuilder.addPredictiveUnits(predictiveUnitDef);
            }

            PredictorDef predictorDef = predictorDefBuilder.build();
            deploymentDefBuiler.setPredictor(predictorDef);

        }

        { // canary testing
            deploymentDefBuiler.setPredictorCanary(predictorDefBuilder.setId("100").build());
        }

        DeploymentDef deploymentDef = deploymentDefBuiler.build();

        return deploymentDef;
    }

    public static DeploymentDef buildExampleDeploymentDef2() {
        DeploymentDef.Builder deploymentDefBuiler = DeploymentDef.newBuilder();

        deploymentDefBuiler.setCluster(ClusterDef.newBuilder());
        deploymentDefBuiler.setId("1");
        deploymentDefBuiler.setName("my deployment");
        deploymentDefBuiler.setUniqueName("my-interesting-project1.my-deployment.1");

        {
            PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();

            predictorDefBuilder.setEnabled(true);
            predictorDefBuilder.setId("0");
            predictorDefBuilder.setName("my_fantastic_predictor");

            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren("1").addChildren("2");

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpu("5")
                        .setGpu("0")
                        .setId("2")
                        .setImage("gsunner/putest")
                        .setMemory("20Gi")
                        .setVersion("")
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
            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren("1").addChildren("2");

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpu("5")
                        .setGpu("0")
                        .setId("2")
                        .setImage("gsunner/putest")
                        .setMemory("20Gi")
                        .setVersion("")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setType(EndpointDef.EndpointType.REST)
                        );
                //@formatter:on

                predictiveUnitDefBuilder.setId("8");

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

        return deploymentDef;
    }
}
