package io.seldon.apife.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import org.springframework.stereotype.Component;

@Component
public class CombinerUnit extends PredictiveUnitBean{
	
	public CombinerUnit() {
		super();
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictorData request, PredictiveUnitState state){
		return new ArrayList<PredictiveUnitState>(state.children.values());
	}

}
