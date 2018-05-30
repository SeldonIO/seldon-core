package io.seldon.wrapper.api;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnExpression;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.SeldonMessageList;
import io.seldon.wrapper.exception.APIException;
import io.seldon.wrapper.exception.APIException.ApiExceptionType;
import io.seldon.wrapper.pb.ProtoBufUtils;

@RestController
@ConditionalOnExpression("${seldon.api.combiner.enabled:false}")
public class CombinerRestController {

	private static Logger logger = LoggerFactory.getLogger(RouterRestController.class.getName());

	@Autowired
	SeldonPredictionService predictionService;

	@RequestMapping(value = "/aggregate", method = {RequestMethod.GET, RequestMethod.POST}, produces = "application/json; charset=utf-8")
    public ResponseEntity<String> route( @RequestParam("json") String json)
	{
		SeldonMessageList request;
		try
		{
			SeldonMessageList.Builder builder = SeldonMessageList.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, json );
			request = builder.build();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE,json);
		}

		try
		{
			SeldonMessage response = predictionService.aggregate(request);
			String res = ProtoBufUtils.toJson(response);
			return new ResponseEntity<String>(res,HttpStatus.OK);
		}
		catch (InvalidProtocolBufferException e) {
			throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE,"");
		} 
	}
}
	