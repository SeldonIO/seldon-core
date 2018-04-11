package io.seldon.example.model;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.wrapper.api.SeldonPredictionService;

@Component
public class ExampleModelHandler implements SeldonPredictionService {

	@Override
	public SeldonMessage predict(SeldonMessage payload) {
		return SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addShape(1).addValues(1.0))).build();
	}

}
