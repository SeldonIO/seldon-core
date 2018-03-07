/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.predictors;

import java.util.Arrays;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.Status;
import io.seldon.protos.PredictionProtos.Tensor;

@Component
public class SimpleModelUnit extends PredictiveUnitImpl {
	
	public SimpleModelUnit() {}

	public static final Double[] values = {0.1,0.9,0.5};		
	public static final String[] classes = {"class0","class1","class2"};
	
	@Override
	public SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state){
		SeldonMessage output = SeldonMessage.newBuilder()
				.setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setMeta(Meta.newBuilder())//.addModel(state.id))
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(classes))
					.setTensor(Tensor.newBuilder().addShape(1).addShape(values.length)
					.addAllValues(Arrays.asList(values)))).build();
		System.out.println("Model " + state.name + " finishing computations");
		return output;
	}
}
