package io.seldon.clustermanager.example;

import io.seldon.protos.DeploymentProtos.ClusterDef;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class KubernetesManagerExampleUtils {

    public static DeploymentDef buildExampleDeploymentDef() {
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

                predictiveUnitDefBuilder.addChildren(1).addChildren(2);

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpus(5)
                        .setImagePullSecret("")
                        .setGpus(0)
                        .setId("2")
                        .setImage("nginx")
                        .setMemoryGb(20)
                        .setReplicas(3)
                        .setVersion("1.9.0")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setContainerPort(80)
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

                predictiveUnitDefBuilder.setSubtype("simpleModel");

                predictiveUnitDefBuilder.setType("model");

                PredictiveUnitDef predictiveUnitDef = predictiveUnitDefBuilder.build();
                predictorDefBuilder.addPredictiveUnits(predictiveUnitDef);
            }
            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren(1).addChildren(2);

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpus(5)
                        .setImagePullSecret("my-registry-secret")
                        .setGpus(0)
                        .setId("2")
                        .setImage("nginx")
                        .setMemoryGb(20)
                        .setReplicas(3)
                        .setVersion("1.9.0")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setContainerPort(80)
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

                predictiveUnitDefBuilder.setSubtype("simpleModel");

                predictiveUnitDefBuilder.setType("model");

                PredictiveUnitDef predictiveUnitDef = predictiveUnitDefBuilder.build();
                predictorDefBuilder.addPredictiveUnits(predictiveUnitDef);
            }

            PredictorDef predictorDef = predictorDefBuilder.build();

            deploymentDefBuiler.setPredictor(predictorDef);

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

                predictiveUnitDefBuilder.addChildren(1).addChildren(2);

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpus(5)
                        .setImagePullSecret("my-registry-secret")
                        .setGpus(0)
                        .setId("2")
                        .setImage("nginx")
                        .setMemoryGb(20)
                        .setReplicas(3)
                        .setVersion("1.9.2")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setContainerPort(80)
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

                predictiveUnitDefBuilder.setSubtype("simpleModel");

                predictiveUnitDefBuilder.setType("model");

                PredictiveUnitDef predictiveUnitDef = predictiveUnitDefBuilder.build();
                predictorDefBuilder.addPredictiveUnits(predictiveUnitDef);
            }
            {
                PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();

                predictiveUnitDefBuilder.addChildren(1).addChildren(2);

                //@formatter:off
                predictiveUnitDefBuilder.setClusterResources(ClusterResourcesDef.newBuilder()
                        .setCpus(5)
                        .setImagePullSecret("my-registry-secret")
                        .setGpus(0)
                        .setId("2")
                        .setImage("nginx")
                        .setMemoryGb(20)
                        .setReplicas(3)
                        .setVersion("1.9.2")
                        );
                //@formatter:on

                //@formatter:off
                predictiveUnitDefBuilder.setEndpoint(EndpointDef.newBuilder()
                        .setServiceHost("127.0.0.1")
                        .setServicePort(5004)
                        .setContainerPort(80)
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

                predictiveUnitDefBuilder.setSubtype("simpleModel");

                predictiveUnitDefBuilder.setType("model");

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
