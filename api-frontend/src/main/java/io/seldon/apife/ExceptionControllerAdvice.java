package io.seldon.apife;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.apife.pb.ProtoBufUtils;
import io.seldon.protos.PredictionProtos.Status;


@ControllerAdvice
public class ExceptionControllerAdvice {


	@ExceptionHandler(SeldonAPIException.class)
	public ResponseEntity<String> handleUnauthorizedException(SeldonAPIException exception) throws InvalidProtocolBufferException {

		
		Status.Builder statusBuilder = Status.newBuilder();
		statusBuilder.setCode(exception.getApiExceptionType().getId());
		statusBuilder.setReason(exception.getApiExceptionType().getMessage());
		statusBuilder.setInfo(exception.getInfo());
		statusBuilder.setStatus(Status.StatusFlag.FAILURE);		
		
		Status status = statusBuilder.build();
		String json;
		json = ProtoBufUtils.toJson(status);
		return new ResponseEntity<String>(json,HttpStatus.valueOf(exception.getApiExceptionType().getHttpCode())); 


	}

	
}