package io.seldon.example.model;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.wrapper.api.model.SeldonModelHandler;

@Component
public class ExampleModelHandler implements SeldonModelHandler {

	@Override
	public SeldonMessage predict(SeldonMessage payload) {
		// TODO Auto-generated method stub
		return SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addShape(1).addValues(1.0))).build();
	}

}
