package io.seldon.wrapper.utils;

import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import org.dmg.pmml.FieldName;
import org.dmg.pmml.MiningFunction;
import org.dmg.pmml.PMML;
import org.jpmml.evaluator.Evaluator;
import org.jpmml.evaluator.FieldValue;
import org.jpmml.evaluator.InputField;
import org.jpmml.evaluator.ModelEvaluatorFactory;
import org.jpmml.evaluator.ProbabilityDistribution;
import org.jpmml.evaluator.ReportingValueFactoryFactory;
import org.jpmml.evaluator.TargetField;
import org.jpmml.evaluator.ValueFactoryFactory;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

import hex.genmodel.easy.RowData;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.DefaultData.DataOneofCase;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.wrapper.exception.APIException;
import io.seldon.wrapper.exception.APIException.ApiExceptionType;

public class PMMLUtils {

	
	public static abstract class  ValuesBuilder {
		public abstract void addValues(double v);
	}
	
	public static class TensorBuilder extends ValuesBuilder {
		Tensor.Builder b;
		
		public TensorBuilder(Tensor.Builder b) {
			super();
			this.b = b;
		}

		@Override
		public void addValues(double v) {
			b.addValues(v);
		}
	}

	public static class ListBuilder extends ValuesBuilder {
		ListValue.Builder b;
		
		public ListBuilder(ListValue.Builder b) {
			super();
			this.b = b;
		}

		@Override
		public void addValues(double v) {
			b.addValues(Value.newBuilder().setNumberValue(v));
		}
	}
	
	private Map<String,InputField> getInputFieldMap(Evaluator evaluator)
	{
		Map<String,InputField> fieldMap = new HashMap<>();
		List<InputField> inputFields = evaluator.getInputFields();
        for(InputField inputField : inputFields){
        	FieldName inputFieldName = inputField.getName();
        	fieldMap.put(inputFieldName.getValue(), inputField);
        }
        return fieldMap;
	}
	

	private void evaluateForTensor(Evaluator evaluator,Map<FieldName, FieldValue> arguments,ValuesBuilder tBuilder)
	{
		Map<FieldName, ?> result = evaluator.evaluate(arguments);
		List<TargetField> targetFields = evaluator.getTargetFields();
		MiningFunction miningFunction = evaluator.getMiningFunction();
		TargetField targetField = targetFields.get(0); 
		switch(miningFunction){
		case CLASSIFICATION:
			FieldName targetFieldName = targetField.getName();
        	Object targetFieldValue = result.get(targetFieldName);
        	if (targetFieldValue instanceof ProbabilityDistribution)
        	{
        		ProbabilityDistribution prob = (ProbabilityDistribution) targetFieldValue;
        		Set<String> categories = prob.getCategories();
        		for (String k : categories)
        		{
        			Double v = prob.getValue(k);
        			tBuilder.addValues(v);
        		}
        		arguments.clear();
        		return;
        	}
        	else
        		throw new IllegalArgumentException("Expected probability distribution");
		case REGRESSION:        	
		default:
			throw new IllegalArgumentException("Expected a classification model, got " + miningFunction);
		} 
	}
	
	public DefaultData evaluate(PMML model,DefaultData data) {
			
		if (data.getNamesCount() == 0)
		{
			throw new APIException(APIException.ApiExceptionType.WRAPPER_INVALID_MESSAGE, "Data for PMML models must contain names for the fields in the prediction message");
		}
		
		ModelEvaluatorFactory modelEvaluatorFactory = ModelEvaluatorFactory.newInstance();
        ValueFactoryFactory valueFactoryFactory = ReportingValueFactoryFactory.newInstance();
        modelEvaluatorFactory.setValueFactoryFactory(valueFactoryFactory);
		Evaluator evaluator = (Evaluator)modelEvaluatorFactory.newModelEvaluator(model);
		Map<FieldName, FieldValue> arguments = new LinkedHashMap<>();
		Map<String,InputField> fieldMap = getInputFieldMap(evaluator);
		
		if (data.getDataOneofCase() == DataOneofCase.TENSOR) {
			Tensor.Builder tBuilder = Tensor.newBuilder();
			TensorBuilder tb = new TensorBuilder(tBuilder);
			List<Double> valuesList = data.getTensor().getValuesList();
			List<Integer> shapeList = data.getTensor().getShapeList();

			if (shapeList.size() > 2) {
				return null;
			}

			int cols = 0;
			if (shapeList.size() == 1)
				cols = shapeList.get(0);
			else
				cols = shapeList.get(1);
			
			for (int i = 0; i < valuesList.size(); i++) {
				if (i > 0 && i % cols == 0) {
					evaluateForTensor(evaluator, arguments, tb);
				}
				String name = data.getNames(i % cols);
				InputField field = fieldMap.get(name);
				if (field != null)
				{
					Object rawValue = valuesList.get(i);
					FieldValue inputFieldValue = field.prepare(rawValue);
					arguments.put(field.getName(), inputFieldValue);
				}
			}
			evaluateForTensor(evaluator, arguments, tb);
			if (shapeList.size() == 1)
				tBuilder.addShape(tBuilder.getValuesCount());
			else
			{
				int outCols = tBuilder.getValuesCount() / shapeList.get(0);
				tBuilder.addShape(tBuilder.getValuesCount() / outCols).addShape(outCols);
			}
			DefaultData.Builder dataBuilder = DefaultData.newBuilder();
			dataBuilder.setTensor(tBuilder);
			return dataBuilder.build();
		}else if (data.getDataOneofCase() == DataOneofCase.NDARRAY) {
			ListValue list = data.getNdarray();
			int rows = list.getValuesCount();
			int cols = list.getValues(0).getListValue().getValuesCount();

			ListValue.Builder rowsBuilder = ListValue.newBuilder();
			for (int i = 0; i < rows; ++i) {

				ListValue.Builder row = ListValue.newBuilder();
				ListBuilder lb = new ListBuilder(row);
				for (int j = 0; j < cols; j++) {
					String name = data.getNames(j % cols); 
					InputField field = fieldMap.get(name);
					if (field != null)
					{
						Double rawValue = list.getValues(i).getListValue().getValues(j).getNumberValue();
						FieldValue inputFieldValue = field.prepare(rawValue);
						arguments.put(field.getName(), inputFieldValue);
					}
				}
				evaluateForTensor(evaluator, arguments, lb);
				rowsBuilder.addValues(Value.newBuilder().setListValue(row.build()));
			}
			DefaultData.Builder dataBuilder = DefaultData.newBuilder();
			dataBuilder.setNdarray(rowsBuilder.build());
			return dataBuilder.build();
		} 
		else
			throw new UnsupportedOperationException("Only Tensor or NDArray is supported");
	}
	
}
