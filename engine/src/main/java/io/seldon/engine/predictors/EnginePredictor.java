package io.seldon.engine.predictors;

import java.util.Base64;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;
import com.google.protobuf.util.JsonFormat.Printer;

import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class EnginePredictor {

    private final static Logger logger = LoggerFactory.getLogger(EnginePredictor.class);
    private final static String ENGINE_PREDICTOR_KEY = "ENGINE_PREDICTOR";

    private PredictorDef predictorDef = null;

    public void init() throws Exception {
        logger.info("init");

        { // setup the predictorDef using the env vars
            String enginePredictorBase64Encoded = System.getenv().get(ENGINE_PREDICTOR_KEY);
            if (enginePredictorBase64Encoded == null) {
                logger.error("FAILED to find env var [{}], will use defaults for engine predictor", ENGINE_PREDICTOR_KEY);
                predictorDef = buildDefaultPredictorDef();
            } else {
                logger.info("FOUND env var [{}], will use for engine predictor", ENGINE_PREDICTOR_KEY);
                byte[] enginePredictorBytes = Base64.getDecoder().decode(enginePredictorBase64Encoded);
                String enginePredictorJson = new String(enginePredictorBytes);
                PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();
                try {
                    updateMessageBuilderFromJson(predictorDefBuilder, enginePredictorJson);
                } catch (Exception e) {
                    logger.error("FAILED extracting predictorDef from env var [{}]", ENGINE_PREDICTOR_KEY,e);
                    throw e;
                }
                predictorDef = predictorDefBuilder.build();
            }
        }

        logger.info("Installed engine predictor: {}", toJson(predictorDef, true));
    }

    public void cleanup() throws Exception {
        logger.info("cleanup");
    }

    public PredictorDef getPredictorDef() {
        return predictorDef;
    }

    private static PredictorDef buildDefaultPredictorDef() {

        //@formatter:off
        PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder()
                .setName("basic-predictor")
                .setRoot("0");
        //@formatter:on

        { // Add predictiveUnit
            //@formatter:off
            PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder()
                    .setId("0")
                    .setName("basic-pu")
                    .setType(PredictiveUnitDef.PredictiveUnitType.MODEL)
                    .setSubtype(PredictiveUnitDef.PredictiveUnitSubType.MODEL_SIMPLEMODEL);
            //@formatter:on

            predictorDefBuilder.addPredictiveUnits(predictiveUnitDefBuilder);
        }
        return predictorDefBuilder.build();
    }

    private static String toJson(Message message, boolean omittingInsignificantWhitespace) throws InvalidProtocolBufferException {
        Printer jsonPrinter = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames();
        if (omittingInsignificantWhitespace) {
            jsonPrinter = jsonPrinter.omittingInsignificantWhitespace();
        }
        return jsonPrinter.print(message);
    }

    private static <T extends Message.Builder> void updateMessageBuilderFromJson(T messageBuilder, String json) throws InvalidProtocolBufferException {
        JsonFormat.parser().ignoringUnknownFields().merge(json, messageBuilder);
    }

}
