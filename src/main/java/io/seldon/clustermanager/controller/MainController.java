package io.seldon.clustermanager.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;

@RestController
public class MainController {

    @Autowired
    private ClusterManager clusterManager;

    @RequestMapping(value = "/ping", method = RequestMethod.GET)
    public String ping() {
        return "pong";
    }

    @RequestMapping(value = "/namespaces", method = RequestMethod.GET)
    public ResponseEntity<String> get_namespaces() {

        CMResultDef cmResultDef = clusterManager.getNamespaces();
        return cmResultDefToResponseEntity(cmResultDef);
    }

    private static ResponseEntity<String> cmResultDefToResponseEntity(CMResultDef cmResultDef) {

        HttpStatus httpStatus = HttpStatus.valueOf(cmResultDef.getCmstatus().getCode());
        String json = null;
        try {
            json = ProtoBufUtils.toJson(cmResultDef);
        } catch (InvalidProtocolBufferException e) {
            httpStatus = HttpStatus.INTERNAL_SERVER_ERROR;
            json = "Error writing json";
        }

        HttpHeaders responseHeaders = new HttpHeaders();
        responseHeaders.setContentType(MediaType.APPLICATION_JSON);
        ResponseEntity<String> responseEntity = new ResponseEntity<String>(json, responseHeaders, httpStatus);

        return responseEntity;
    }
}
