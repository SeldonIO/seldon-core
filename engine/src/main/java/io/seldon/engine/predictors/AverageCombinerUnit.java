package io.seldon.engine.predictors;

import java.util.Iterator;
import java.util.List;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.nd4j.linalg.factory.Nd4j;
import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;

import io.seldon.engine.predictors.PredictorUtils;

@Component
public class AverageCombinerUnit extends CombinerUnit{
	
	public AverageCombinerUnit() {}

	@Override
	public SeldonMessage aggregateOutputs(List<SeldonMessage> outputs, PredictiveUnitState state){
		
		if (outputs.size()==0){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner received no inputs"));
		}
		
		int[] shape = PredictorUtils.getShape(outputs.get(0).getData());
		
		if (shape == null){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner cannot extract data shape"));
		}
		
		if (shape.length!=2){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner received data that is not 2 dimensional"));
		}
		
		INDArray currentSum = Nd4j.zeros(shape[0],shape[1]);
		SeldonMessage.Builder respBuilder = SeldonMessage.newBuilder();
		
		for (Iterator<SeldonMessage> i = outputs.iterator(); i.hasNext();)
		{
			DefaultData inputData = i.next().getData();
			int[] inputShape = PredictorUtils.getShape(inputData);
			if (inputShape == null){
				throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner cannot extract data shape"));
			}
			if (inputShape.length!=2){
				throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner received data that is not 2 dimensional"));
			}
			if (inputShape[0] != shape[0]){
				throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Expected batch length %d but found %d",shape[0],inputShape[0]));
			}
			if (inputShape[1] != shape[1]){
				throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Expected batch length %d but found %d",shape[1],inputShape[1]));
			}
			INDArray inputArr = PredictorUtils.getINDArray(inputData);
			currentSum = currentSum.add(inputArr);
		}
		currentSum = currentSum.div((float)outputs.size());
		
		DefaultData newData = PredictorUtils.updateData(outputs.get(0).getData(), currentSum);
		respBuilder.setData(newData);
		respBuilder.setMeta(outputs.get(0).getMeta());
		respBuilder.setStatus(outputs.get(0).getStatus());
		
		return respBuilder.build();
	}

}
