package io.seldon.engine;

import com.google.protobuf.util.JsonFormat;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.exception.APIException.ApiExceptionType;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Status;
import org.junit.Assert;
import org.junit.Test;
import org.springframework.http.ResponseEntity;

public class ExceptionControllerAdviceTest {
    @Test
    public void testApiExceptionType() throws Exception {
        ResponseEntity<String> responseEntity = new io.seldon.engine.ExceptionControllerAdvice()
                .handleUnauthorizedException(new APIException(
                        ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "info"));
        validateSeldonMessage(responseEntity, ApiExceptionType.ENGINE_MICROSERVICE_ERROR);
    }

    @Test
    public void testCustomizedExceptionType() throws Exception {
        ApiExceptionType exceptionType =
                ApiExceptionType.CUSTOMIZED_EXCEPTION;
        exceptionType.setMessage("exception msg in test");
        ResponseEntity<String> responseEntity = new io.seldon.engine.ExceptionControllerAdvice()
                .handleUnauthorizedException(new APIException(exceptionType, "info"));
        validateSeldonMessage(responseEntity, exceptionType);
    }

    private void validateSeldonMessage(
            ResponseEntity<String> httpResponse, ApiExceptionType exceptionType) throws Exception {
        String response = httpResponse.getBody();
        SeldonMessage.Builder builder = SeldonMessage.newBuilder();
        JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
        SeldonMessage seldonMessage = builder.build();

        Assert.assertEquals(exceptionType.getHttpCode(), httpResponse.getStatusCodeValue());
        Assert.assertEquals(exceptionType.getId(), seldonMessage.getStatus().getCode());
        Assert.assertEquals(exceptionType.getMessage(), seldonMessage.getStatus().getReason());
        Assert.assertEquals("info", seldonMessage.getStatus().getInfo());
        Assert.assertEquals(Status.StatusFlag.FAILURE, seldonMessage.getStatus().getStatus());
    }
}
