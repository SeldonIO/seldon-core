package io.seldon.engine.predictors;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Base64;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;

import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.Time;
import io.kubernetes.client.proto.Meta.Timestamp;
import io.kubernetes.client.proto.Resource.Quantity;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.engine.pb.IntOrStringUtils;
import io.seldon.engine.pb.JsonFormat;
import io.seldon.engine.pb.JsonFormat.Printer;
import io.seldon.engine.pb.QuantityUtils;
import io.seldon.engine.pb.TimeUtils;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitSubtype;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictorSpec;



public class EnginePredictor {

    private final static Logger logger = LoggerFactory.getLogger(EnginePredictor.class);
    private final static String ENGINE_PREDICTOR_KEY = "ENGINE_PREDICTOR";

    private PredictorSpec predictorSpec = null;

    public void init() throws Exception {
        logger.info("init");

        { // setup the PredictorSpec using the env vars
            String enginePredictorBase64Encoded = System.getenv().get(ENGINE_PREDICTOR_KEY);
            if (enginePredictorBase64Encoded == null) {
            	String filePath = "./deploymentdef.json";
            	File deploymentFile = new File(filePath);
            	if (deploymentFile.exists()){
            		logger.error("FAILED to find env var [{}], will use json file", ENGINE_PREDICTOR_KEY);
            		byte[] encoded = Files.readAllBytes(Paths.get(filePath));
            		String enginePredictorJson = new String(encoded);
            		PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
	                try {
	                    updateMessageBuilderFromJson(PredictorSpecBuilder, enginePredictorJson);
	                } catch (Exception e) {
	                    logger.error("FAILED building PredictorSpec from file content", ENGINE_PREDICTOR_KEY,e);
	                    throw e;
	                }
	                predictorSpec = PredictorSpecBuilder.build();
            	}
            	else {	
            		logger.error("FAILED to find env var [{}], will use defaults for engine predictor", ENGINE_PREDICTOR_KEY);
            		predictorSpec = buildDefaultPredictorSpec();
            	}
            } else {
                logger.info("FOUND env var [{}], will use for engine predictor", ENGINE_PREDICTOR_KEY);
                byte[] enginePredictorBytes = Base64.getDecoder().decode(enginePredictorBase64Encoded);
                String enginePredictorJson = new String(enginePredictorBytes);
                PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
                try {
                    updateMessageBuilderFromJson(PredictorSpecBuilder, enginePredictorJson);
                } catch (Exception e) {
                    logger.error("FAILED extracting PredictorSpec from env var [{}]", ENGINE_PREDICTOR_KEY,e);
                    throw e;
                }
                predictorSpec = PredictorSpecBuilder.build();
            }
        }

        logger.info("Installed engine predictor: {}", toJson(predictorSpec, true));
    }

    public void cleanup() throws Exception {
        logger.info("cleanup");
    }

    public PredictorSpec getPredictorSpec() {
        return predictorSpec;
    }

    private static PredictorSpec buildDefaultPredictorSpec() {

        //@formatter:off
        PredictorSpec.Builder predictorSpecBuilder = PredictorSpec.newBuilder()
                .setName("basic-predictor")
                .setComponentSpec(PodTemplateSpec.newBuilder());
        //@formatter:on

        { // Add predictorGraph
            //@formatter:off
        	PredictiveUnit.Builder PredictiveUnitBuilder = PredictiveUnit.newBuilder()
        			.setName("basic-pu")
        			.setType(PredictiveUnitType.MODEL)
        			.setSubtype(PredictiveUnitSubtype.SIMPLE_MODEL);
            //@formatter:on

        	predictorSpecBuilder.setGraph(PredictiveUnitBuilder);
        }
        return predictorSpecBuilder.build();
    }

    private static String toJson(Message message, boolean omittingInsignificantWhitespace) throws InvalidProtocolBufferException {
        Printer jsonPrinter = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames();
        if (omittingInsignificantWhitespace) {
            jsonPrinter = jsonPrinter.omittingInsignificantWhitespace();
        }
        return jsonPrinter.print(message);
    }

    private static <T extends Message.Builder> void updateMessageBuilderFromJson(T messageBuilder, String json) throws InvalidProtocolBufferException {
        JsonFormat.parser().ignoringUnknownFields()
        .usingTypeParser(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
        .usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
        .usingTypeParser(Time.getDescriptor().getFullName(), new TimeUtils.TimeParser())
        .usingTypeParser(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeParser()) 
        .merge(json, messageBuilder);
    }

}
