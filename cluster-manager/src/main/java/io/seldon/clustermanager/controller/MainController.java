package io.seldon.clustermanager.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.CMStatusDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

@RestController
public class MainController {

    private final static Logger logger = LoggerFactory.getLogger(MainController.class);

    @Autowired
    private ClusterManager clusterManager;

    @RequestMapping(value = "/ping", method = RequestMethod.GET)
    public String ping() {
        return "pong";
    }

    @RequestMapping(value = "/authping", method = RequestMethod.GET)
    public String authping() {
        return "authpong";
    }

    @RequestMapping(value = "/namespaces", method = RequestMethod.GET)
    public ResponseEntity<String> get_namespaces() {

        CMResultDef cmResultDef = clusterManager.getNamespaces();
        return ControllerUtils.cmResultDefToResponseEntity(cmResultDef);
    }

    @RequestMapping(value = "/deployments", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> deployments_post(RequestEntity<String> requestEntity) {

        String json = requestEntity.getBody();
        logger.debug(String.format("[%s] [%s] [%s]", "POST", requestEntity.getUrl().getPath(), json));

        CMResultDef cmResultDef = null;
        try {
            DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
            ProtoBufUtils.updateMessageBuilderFromJson(deploymentDefBuilder, json);
            cmResultDef = clusterManager.createSeldonDeployment(deploymentDefBuilder.build());
        } catch (InvalidProtocolBufferException e) {
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            //@formatter:off
            cmResultDef = CMResultDef.newBuilder()
                    .setCmstatus(CMStatusDef.newBuilder()
                            .setCode(500)
                            .setStatus(CMStatusDef.Status.FAILURE)
                            .setInfo(info))
                    .clearOneofData()
                    .build();
            //@formatter:on
        }

        return ControllerUtils.cmResultDefToResponseEntity(cmResultDef);
    }

    @RequestMapping(value = "/deployments/{id}", method = RequestMethod.DELETE, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> deployments_delete(@PathVariable("id") String id, RequestEntity<String> requestEntity) {

        String seldon_deployment_id = id;
        logger.debug(String.format("[%s] [%s]", "DELETE", requestEntity.getUrl().getPath()));

        DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
        deploymentDefBuilder.setId(seldon_deployment_id);
        CMResultDef cmResultDef = clusterManager.deleteSeldonDeployment(deploymentDefBuilder.build());

        return ControllerUtils.cmResultDefToResponseEntity(cmResultDef);
    }

    @RequestMapping(value = "/deployments", method = RequestMethod.PATCH, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> deployments_patch(RequestEntity<String> requestEntity) {

        String json = requestEntity.getBody();
        logger.debug(String.format("[%s] [%s] [%s]", "PATCH", requestEntity.getUrl().getPath(), json));

        CMResultDef cmResultDef = null;
        try {
            DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
            ProtoBufUtils.updateMessageBuilderFromJson(deploymentDefBuilder, json);
            cmResultDef = clusterManager.updateSeldonDeployment(deploymentDefBuilder.build());
        } catch (InvalidProtocolBufferException e) {
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            //@formatter:off
            cmResultDef = CMResultDef.newBuilder()
                    .setCmstatus(CMStatusDef.newBuilder()
                            .setCode(500)
                            .setStatus(CMStatusDef.Status.FAILURE)
                            .setInfo(info))
                    .clearOneofData()
                    .build();
            //@formatter:on
        }

        return ControllerUtils.cmResultDefToResponseEntity(cmResultDef);
    }

}
