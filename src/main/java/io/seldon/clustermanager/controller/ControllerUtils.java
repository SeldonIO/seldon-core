package io.seldon.clustermanager.controller;

import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;

public class ControllerUtils {

    public static ResponseEntity<String> cmResultDefToResponseEntity(CMResultDef cmResultDef) {

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
