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

import java.util.List;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

public abstract class PredictiveUnitImpl {
	
	public boolean ready(PredictiveUnitState state) {
		return true;
	}

	public SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return input;
	}
	
	public SeldonMessage transformOutput(SeldonMessage output, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return output;
	}
	
	public SeldonMessage aggregate(List<SeldonMessage> outputs, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return outputs.get(0);
	}
	
	public int route(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return -1;
	}
	
	public void doSendFeedback(Feedback feedback, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return;
	}
}
