package io.seldon.apife.predictors;

import java.io.IOException;
import java.util.List;

import org.springframework.stereotype.Component;

@Component
public class ModelUnit extends PredictiveUnitBean{
	
	public ModelUnit() {
		super();
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictorData request, PredictiveUnitState data){
		return null;
	}
	
	@Override
	protected PredictorReturn backwardPass(List<PredictorData> inputs, PredictiveUnitState state){
		PredictorRequest request = (PredictorRequest) inputs.get(0);
		PredictorReturn ret = null;
		try {
			ret = internalPredictionService.getPrediction(request, state.endpoint);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return ret;
	}

}
