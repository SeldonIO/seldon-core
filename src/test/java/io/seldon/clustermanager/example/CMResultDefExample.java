package io.seldon.clustermanager.example;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.CMStatus;

public class CMResultDefExample {

    public static void main(String[] args) {
        test_failure();
        test_success_with_test_str();
        test_success_with_test_num();
    }

    private static void test_failure() {
        System.out.println("---------- Testing Failure ----------");
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatus.newBuilder()
                        .setCode(500)
                        .setInfo("Something went wrong")
                        .setReason("BadStuff")
                        .setStatus(CMStatus.Status.FAILURE))
                .clearOneofData()
                .build();
        //@formatter:on

        String json = toJson(cmResultDef);
        System.out.println(json);
        CMResultDef cmResultDef2 = fromJson(json);
        examineCMResultDef(cmResultDef2);
    }

    private static void test_success_with_test_str() {
        System.out.println("---------- Testing Success with test_str ----------");
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatus.newBuilder()
                        .setCode(200)
                        .setStatus(CMStatus.Status.SUCCESS))
                .setTestStr("Some String Value")
                .build();
        //@formatter:on

        String json = toJson(cmResultDef);
        System.out.println(json);
        CMResultDef cmResultDef2 = fromJson(json);
        examineCMResultDef(cmResultDef2);
    }

    private static void test_success_with_test_num() {
        System.out.println("---------- Testing Success with test_num ----------");
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatus.newBuilder()
                        .setCode(200)
                        .setStatus(CMStatus.Status.SUCCESS))
                .setTestNum(42)
                .build();
        //@formatter:on

        String json = toJson(cmResultDef);
        System.out.println(json);
        CMResultDef cmResultDef2 = fromJson(json);
        examineCMResultDef(cmResultDef2);
    }

    private static String toJson(CMResultDef cmResultDef) {
        String json = null;
        try {

            json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(cmResultDef);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
        return json;
    }

    private static CMResultDef fromJson(String json) {
        CMResultDef.Builder cmResultDefBuilder = CMResultDef.newBuilder();
        try {
            JsonFormat.parser().ignoringUnknownFields().merge(json, cmResultDefBuilder);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
        return cmResultDefBuilder.build();
    }

    private static void examineCMResultDef(CMResultDef cmResultDef) {
        CMStatus.Status status = cmResultDef.getCmstatus().getStatus();
        if (status == CMStatus.Status.SUCCESS) {
            System.out.println("SUCCESS!");
        } else if (status == CMStatus.Status.FAILURE) {
            System.out.println("FAILURE!");
        } else {
            System.out.println("NOT SUCCESS!");
        }

        switch (cmResultDef.getOneofDataCase()) {
        case ONEOFDATA_NOT_SET:
            System.out.println("Data not set!");
            break;
        case TEST_NUM:
            System.out.println(String.format("The data is test_num[%d]", cmResultDef.getTestNum()));
            break;
        case TEST_STR:
            System.out.println(String.format("The data is test_str[%s]", cmResultDef.getTestStr()));
            break;
        }

    }
}
