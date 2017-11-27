package io.seldon.engine.predictors;

import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.PredictionProtos.DefaultData.DataOneofCase;
import io.seldon.protos.PredictionProtos.DefaultData;

import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.stream.Collectors;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.nd4j.linalg.factory.Nd4j;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

public class PredictorUtils {
	
	public static DefaultData tensorToNDArray(DefaultData data){
		if (data.getDataOneofCase() == DataOneofCase.TENSOR)
		{
			List<Double> valuesList = data.getTensor().getValuesList();
			List<Integer> shapeList = data.getTensor().getShapeList();
			
			DefaultData.Builder dataBuilder = DefaultData.newBuilder();
					
			int index=0;
			for (Iterator<String> i = data.getNamesList().iterator(); i.hasNext();){
				dataBuilder.setNames(index, i.next());
				index++;
			}
			
			ListValue.Builder b1 = ListValue.newBuilder();
			for (int i = 0; i < shapeList.get(0); ++i) {
				ListValue.Builder b2 = ListValue.newBuilder();
				for (int j = 0; j < shapeList.get(1); j++){
					b2.addValues(Value.newBuilder().setNumberValue(valuesList.get(i*shapeList.get(1))+j));
				}
				b1.addValues(Value.newBuilder().setListValue(b2.build()));
			}
			dataBuilder.setNdarray(b1.build());
			
			return dataBuilder.build();
			
		}
		else if (data.getDataOneofCase() == DataOneofCase.NDARRAY)
		{
			return data;
		}
		return null;
	}
	
	public static DefaultData nDArrayToTensor(DefaultData data){
		if (data.getDataOneofCase() == DataOneofCase.TENSOR)
		{
			return data;
			
		}
		else if (data.getDataOneofCase() == DataOneofCase.NDARRAY)
		{
			int bLength = data.getNdarray().getValuesCount();
			int vLength = data.getNdarray().getValues(0).getListValue().getValuesCount();
			
			DefaultData.Builder dataBuilder = DefaultData.newBuilder();
			Tensor.Builder tBuilder = Tensor.newBuilder().addShape(bLength).addShape(vLength);
			
			int index=0;
			for (Iterator<String> i = data.getNamesList().iterator(); i.hasNext();){
				dataBuilder.setNames(index, i.next());
				index++;
			}
			
			for (int i=0; i<bLength; ++i){
				for (int j=0; j<vLength; ++j){
					tBuilder.addValues(data.getNdarray().getValues(i).getListValue().getValues(j).getNumberValue());
				}
			}
			
			dataBuilder.setTensor(tBuilder);
			
			return dataBuilder.build();
		}
		return null;
	}
	
	public static INDArray getINDArray(DefaultData data){
		
		if (data.getDataOneofCase() == DataOneofCase.TENSOR)
		{
			
			List<Double> valuesList = data.getTensor().getValuesList();
			List<Integer> shapeList = data.getTensor().getShapeList();
			
			double[] values = new double[valuesList.size()];
			int[] shape = new int[shapeList.size()];
			for (int i = 0; i < values.length; i++) {
				values[i] = valuesList.get(i);
			 }
			for (int i = 0; i < shape.length; i++) {
				shape[i] = shapeList.get(i);
			 }
			 		 
			INDArray newArr = Nd4j.create(values,shape,'c');
			
			return newArr;
		}
		else if (data.getDataOneofCase() == DataOneofCase.NDARRAY)
		{
			ListValue list = data.getNdarray();
			int bLength = list.getValuesCount();
			int vLength = list.getValues(0).getListValue().getValuesCount();
			
			double[] values = new double[bLength*vLength];
			int[] shape = {bLength,vLength};
			
			for (int i = 0; i < bLength; ++i) {
				for (int j = 0; j < vLength; j++){
					values[i*bLength+j] = list.getValues(i).getListValue().getValues(j).getNumberValue();
				}
			}
			
			INDArray newArr = Nd4j.create(values,shape,'c');
			
			return newArr;
		}
		return null;
	}
	
	public static int[] getShape(DefaultData data){
		if (data.getDataOneofCase() == DataOneofCase.TENSOR){
			List<Integer> shapeList = data.getTensor().getShapeList();
			int[] shape = new int[shapeList.size()];
			for (int i = 0; i < shape.length; ++i){
				shape[i] = shapeList.get(i);
			}
			
			return shape;
		}
		else if(data.getDataOneofCase() == DataOneofCase.NDARRAY){
			int bLength = data.getNdarray().getValuesCount();
			int vLength = data.getNdarray().getValues(0).getListValue().getValuesCount();
			
			int[] shape = {bLength,vLength};
			return shape;
		}
		return null;
	}
	
	public static DefaultData updateData(DefaultData oldData, INDArray newData){
		DefaultData.Builder dataBuilder = DefaultData.newBuilder();
		
		dataBuilder.addAllNames(oldData.getNamesList());
		
//		int index=0;
//		for (Iterator<String> i = oldData.getFeaturesList().iterator(); i.hasNext();){
//			dataBuilder.setFeatures(index, i.next());
//			index++;
//		}
		
		if (oldData.getDataOneofCase() == DataOneofCase.TENSOR){
			Tensor.Builder tBuilder = Tensor.newBuilder();
			List<Integer> shapeList = Arrays.stream(newData.shape()).boxed().collect(Collectors.toList());
			tBuilder.addAllShape(shapeList);
			
			for (int i=0; i<shapeList.get(0); ++i){
				for (int j=0; j<shapeList.get(1); ++j){
					tBuilder.addValues(newData.getDouble(i,j));
				}
			}
			dataBuilder.setTensor(tBuilder);
			return dataBuilder.build();
		}
		else if (oldData.getDataOneofCase() == DataOneofCase.NDARRAY){
			ListValue.Builder b1 = ListValue.newBuilder();
			for (int i = 0; i < newData.shape()[0]; ++i) {
				ListValue.Builder b2 = ListValue.newBuilder();
				for (int j = 0; j < newData.shape()[1]; j++){
					b2.addValues(Value.newBuilder().setNumberValue(newData.getDouble(i,j)));
				}
				b1.addValues(Value.newBuilder().setListValue(b2.build()));
			}
			dataBuilder.setNdarray(b1.build());
			return dataBuilder.build();
		}
		return null;
		
	}

}

