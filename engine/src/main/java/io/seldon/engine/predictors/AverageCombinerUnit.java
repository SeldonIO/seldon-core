package io.seldon.engine.predictors;

import java.util.Arrays;
import java.util.Iterator;
import java.util.List;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.nd4j.linalg.factory.Nd4j;
import org.springframework.stereotype.Component;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.DefaultDataDef.DataOneofCase;
import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.protos.PredictionProtos.Tensor;

import io.seldon.engine.predictors.PredictorUtils;

@Component
public class AverageCombinerUnit extends CombinerUnit{
	
	public AverageCombinerUnit() {}

	@Override
	public ResponseDef backwardPass(List<ResponseDef> inputs, RequestDef request, PredictiveUnitState state){
		
		if (inputs.size()==0){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner received no inputs"));
		}
		
		int[] shape = PredictorUtils.getShape(inputs.get(0).getResponse());
		
		if (shape == null){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner cannot extract data shape"));
		}
		
		if (shape.length!=2){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Combiner received data that is not 2 dimensional"));
		}
		
		INDArray currentSum = Nd4j.zeros(shape[0],shape[1]);
		ResponseDef.Builder respBuilder = ResponseDef.newBuilder();
		
		for (Iterator<ResponseDef> i = inputs.iterator(); i.hasNext();)
		{
			DefaultDataDef inputData = i.next().getResponse();
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
		currentSum = currentSum.div((float)inputs.size());
		
		DefaultDataDef newData = PredictorUtils.updateData(inputs.get(0).getResponse(), currentSum);
		respBuilder.setResponse(newData);
		respBuilder.setMeta(inputs.get(0).getMeta());
		respBuilder.setStatus(inputs.get(0).getStatus());
		
		return respBuilder.build();
	}
	
	public ResponseDef backwardPassOld(List<ResponseDef> inputs, RequestDef request, PredictiveUnitState state){
		
		Integer batchLength = 0;
		Integer valuesLength = 0;
		Integer inputsLength = inputs.size();
		Boolean initialised = false;
		Double[] averages = null;
		DataOneofCase dataType = DataOneofCase.DATAONEOF_NOT_SET;
		
		ResponseDef.Builder respBuilder = ResponseDef.newBuilder();
		MetaDef.Builder metaBuilder = MetaDef.newBuilder();
		DefaultDataDef.Builder dataBuilder = DefaultDataDef.newBuilder();
		for (ResponseDef predRet : inputs){
//			metaBuilder.addAllModel(predRet.getMeta().getModelList());
			int bLength = 0;
			int vLength = 0;
			if (predRet.getResponse().getDataOneofCase() == DataOneofCase.TENSOR)
			{
				Tensor tensor = predRet.getResponse().getTensor();
				if (tensor.getShapeCount() == 2)
				{
					bLength = tensor.getShape(0);
					vLength = tensor.getShape(1);
				}
				else
				{
					bLength = 1;
					vLength = tensor.getValuesCount();
				}
			}
			else if (predRet.getResponse().getDataOneofCase() == DataOneofCase.NDARRAY)// nDArray
			{
				ListValue list = predRet.getResponse().getNdarray();
				bLength = list.getValuesCount();
				vLength = list.getValues(0).getListValue().getValuesCount();
			}
			if (!initialised){
				dataType = predRet.getResponse().getDataOneofCase();
				batchLength = bLength;
				valuesLength = vLength;
				averages = new Double[batchLength*valuesLength];
				Arrays.fill(averages, 0.);
				respBuilder.setMeta(predRet.getMeta()).setStatus(predRet.getStatus());
				dataBuilder.addAllNames(predRet.getResponse().getNamesList());
				initialised = true;
			}
			else
			{
				if (bLength != batchLength)
				{
					throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Expected batch length %d but found %d",batchLength,bLength));
				}
				if (vLength != valuesLength)
				{
					throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Expected values length %d but found %d",valuesLength,vLength));
				}
			}
			for (int i = 0; i < batchLength; ++i) {
				for (int j = 0; j < valuesLength; j++){
					if (predRet.getResponse().getDataOneofCase() == DataOneofCase.TENSOR)
						averages[(i*valuesLength)+j] += predRet.getResponse().getTensor().getValues((i*valuesLength)+j);
					else if (predRet.getResponse().getDataOneofCase() == DataOneofCase.NDARRAY)
						averages[(i*valuesLength)+j] += predRet.getResponse().getNdarray().getValues(i).getListValue().getValues(j).getNumberValue();
				}
			}
		}
		
		for (int i = 0; i < batchLength; ++i) {
			for (int j = 0; j < valuesLength; j++){
				averages[(i*valuesLength)+j] /= inputsLength;
			}
		}
	
		if (averages != null)
		{
			if (dataType == DataOneofCase.TENSOR)
			{
				dataBuilder.setTensor(Tensor.newBuilder().addShape(batchLength).addShape(valuesLength).addAllValues(Arrays.asList(averages)).build());
			}
			else if(dataType == DataOneofCase.NDARRAY)
			{
				ListValue.Builder b1 = ListValue.newBuilder();
				for (int i = 0; i < batchLength; ++i) {
					ListValue.Builder b2 = ListValue.newBuilder();
					for (int j = 0; j < valuesLength; j++){
						b2.addValues(Value.newBuilder().setNumberValue(averages[(i*valuesLength)+j]));
					}
					b1.addValues(Value.newBuilder().setListValue(b2.build()));
				}
				dataBuilder.setNdarray(b1.build());
			}
		}
		respBuilder.setResponse(dataBuilder).setMeta(metaBuilder);
		
		return respBuilder.build();
	}

}
