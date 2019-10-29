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
        APIException exception = new APIException(
                ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "info");
        ResponseEntity<String> responseEntity = new io.seldon.engine.ExceptionControllerAdvice()
                .handleUnauthorizedException(exception);
        validateSeldonMessage(responseEntity, exception);
    }

    @Test
    public void testCustomizedExceptionType() throws Exception {
        APIException exception = new APIException(400,
                                                 "test in message",
                                                 200,
                                                 "info");
        ResponseEntity<String> responseEntity = new io.seldon.engine.ExceptionControllerAdvice()
                .handleUnauthorizedException(exception);
        validateSeldonMessage(responseEntity, exception);
    }

    private void validateSeldonMessage(
            ResponseEntity<String> httpResponse, APIException exception) throws Exception {
        String response = httpResponse.getBody();
        SeldonMessage.Builder builder = SeldonMessage.newBuilder();
        JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
        SeldonMessage seldonMessage = builder.build();

        Assert.assertEquals(exception.getHttpCode(), httpResponse.getStatusCodeValue());
        Assert.assertEquals(exception.getId(), seldonMessage.getStatus().getCode());
        Assert.assertEquals(exception.getMessage(), seldonMessage.getStatus().getReason());
        Assert.assertEquals("info", seldonMessage.getStatus().getInfo());
        Assert.assertEquals(Status.StatusFlag.FAILURE, seldonMessage.getStatus().getStatus());
    }
}
