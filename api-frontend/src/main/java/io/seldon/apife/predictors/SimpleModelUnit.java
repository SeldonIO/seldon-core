package io.seldon.apife.predictors;

import java.util.List;

import org.springframework.stereotype.Component;

@Component
public class SimpleModelUnit extends ModelUnit {
	
	public SimpleModelUnit() {}

	@Override
	protected PredictorReturn backwardPass(List<PredictorData> inputs, PredictiveUnitState state){
		System.out.println("Model " + state.name + " starting computations");
		String[] classes = {"class0","class1","class2"};
		Double[][] values = {{0.1,0.9,0.5}};
		PredictorReturn ret = new PredictorReturn(classes, values);
		try {
			Thread.sleep(20);
		} catch (InterruptedException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		System.out.println("Model " + state.name + " finishing computations");
		return ret;
	}
}
