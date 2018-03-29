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

import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.proto.V1.Container;
import io.kubernetes.client.proto.V1.PodSpec;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@RunWith(SpringRunner.class)
@SpringBootTest
public class SimpleModelUnitTest {

	@Autowired
	PredictorBean predictorBean;
	
	@Test
	public void simpleTest() throws InterruptedException, ExecutionException, InvalidProtocolBufferException
	{
		PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
		// PredictorSpecBuilder.setEnabled(true);
		
		PredictorSpecBuilder.setName("p1");
		// PredictorSpecBuilder.setRoot("1");
		PredictorSpecBuilder.setReplicas(1);
		PredictorSpecBuilder.setComponentSpec(PodTemplateSpec.newBuilder());
		
		PredictiveUnit.Builder PredictiveUnitBuilder = PredictiveUnit.newBuilder();
		PredictiveUnitBuilder.setName("1");
		PredictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		PredictiveUnitBuilder.setImplementation(PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
		
		PredictorSpecBuilder.setGraph(PredictiveUnitBuilder.build());
		PredictorSpec predictor = PredictorSpecBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

		
		SeldonMessage p = SeldonMessage.newBuilder().build();

		
        SeldonMessage predictorReturn = predictorBean.predict(p,predictorState);
        
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2),0);   
		
        
	}
	
	
	@Test
	public void simpleTestWithImageNoVersion() throws InterruptedException, ExecutionException, InvalidProtocolBufferException
	{
		PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
		// PredictorSpecBuilder.setEnabled(true);
		
		PredictorSpecBuilder.setName("p1");
		// PredictorSpecBuilder.setRoot("1");
		PredictorSpecBuilder.setReplicas(1);
		
		final String imageName = "myimage";
		PodTemplateSpec.Builder ptsBuilder = PodTemplateSpec.newBuilder().setSpec(PodSpec.newBuilder().addContainers(Container.newBuilder().setImage(imageName).setName("1")));
		
		PredictorSpecBuilder.setComponentSpec(ptsBuilder);
		
		PredictiveUnit.Builder PredictiveUnitBuilder = PredictiveUnit.newBuilder();
		PredictiveUnitBuilder.setName("1");
		PredictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		PredictiveUnitBuilder.setImplementation(PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
		
		PredictorSpecBuilder.setGraph(PredictiveUnitBuilder.build());
		PredictorSpec predictor = PredictorSpecBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

		Assert.assertEquals(imageName, predictorState.getRootState().imageName);
		Assert.assertEquals("", predictorState.getRootState().imageVersion);
		
		SeldonMessage p = SeldonMessage.newBuilder().build();

		
        SeldonMessage predictorReturn = predictorBean.predict(p,predictorState);
        
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2),0);   
		
        
	}
	
	@Test
	public void simpleTestWithImageVersion() throws InterruptedException, ExecutionException, InvalidProtocolBufferException
	{
		PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
		// PredictorSpecBuilder.setEnabled(true);
		
		PredictorSpecBuilder.setName("p1");
		// PredictorSpecBuilder.setRoot("1");
		PredictorSpecBuilder.setReplicas(1);
		
		final String imageName = "myimage";
		final String imageVersion = "0.1";
		PodTemplateSpec.Builder ptsBuilder = PodTemplateSpec.newBuilder().setSpec(PodSpec.newBuilder().addContainers(Container.newBuilder().setImage(imageName+":"+imageVersion).setName("1")));
		
		PredictorSpecBuilder.setComponentSpec(ptsBuilder);
		
		PredictiveUnit.Builder PredictiveUnitBuilder = PredictiveUnit.newBuilder();
		PredictiveUnitBuilder.setName("1");
		PredictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		PredictiveUnitBuilder.setImplementation(PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
		
		PredictorSpecBuilder.setGraph(PredictiveUnitBuilder.build());
		PredictorSpec predictor = PredictorSpecBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

		Assert.assertEquals(imageName, predictorState.getRootState().imageName);
		Assert.assertEquals(imageVersion, predictorState.getRootState().imageVersion);
		
		SeldonMessage p = SeldonMessage.newBuilder().build();

		
        SeldonMessage predictorReturn = predictorBean.predict(p,predictorState);
        
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2),0);   
		
        
	}
	
}
