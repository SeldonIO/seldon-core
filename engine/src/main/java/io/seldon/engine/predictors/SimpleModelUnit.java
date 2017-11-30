package io.seldon.engine.predictors;

import java.util.Arrays;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.Status;
import io.seldon.protos.PredictionProtos.Tensor;

@Component
public class SimpleModelUnit extends ModelUnit {
	
	public SimpleModelUnit() {}

	public static final Double[] values = {0.1,0.9,0.5};		
	public static final String[] classes = {"class0","class1","class2"};
	
	@Override
	protected Message transformInput(Message input, PredictiveUnitState state){
		Message output = Message.newBuilder()
				.setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setMeta(Meta.newBuilder())//.addModel(state.id))
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(classes))
					.setTensor(Tensor.newBuilder().addShape(1).addShape(values.length)
					.addAllValues(Arrays.asList(values)))).build();
		try {
			Thread.sleep(20);
		} catch (InterruptedException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		System.out.println("Model " + state.name + " finishing computations");
		return output;
	}
}
