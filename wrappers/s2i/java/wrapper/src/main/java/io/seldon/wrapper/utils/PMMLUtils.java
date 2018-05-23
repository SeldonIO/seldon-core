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

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.DefaultData.DataOneofCase;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.wrapper.exception.APIException;
import io.seldon.wrapper.exception.APIException.ApiExceptionType;

public class PMMLUtils {

	
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
	
	private void evaluateForTensor(Evaluator evaluator,Map<FieldName, FieldValue> arguments,Tensor.Builder tBuilder)
	{
		Map<FieldName, ?> result = evaluator.evaluate(arguments);
		List<TargetField> targetFields = evaluator.getTargetFields();
		MiningFunction miningFunction = evaluator.getMiningFunction();
		TargetField targetField = targetFields.get(0); 
		switch(miningFunction){
		case REGRESSION:
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
        		return;
        	}
        	else
        		throw new IllegalArgumentException("Expected probability distribution");
		default:
			throw new IllegalArgumentException("Expected a regression or classification model, got " + miningFunction);
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
					evaluateForTensor(evaluator, arguments, tBuilder);
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
			evaluateForTensor(evaluator, arguments, tBuilder);
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
		}
		else
			throw new UnsupportedOperationException("NDArray not supported at present");
	}
	
}
