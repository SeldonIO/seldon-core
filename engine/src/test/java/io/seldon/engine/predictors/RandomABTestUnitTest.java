package io.seldon.engine.predictors;

import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.protos.DeploymentProtos.Parameter;
import io.seldon.protos.DeploymentProtos.Parameter.ParameterType;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@RunWith(SpringRunner.class)
@SpringBootTest
public class RandomABTestUnitTest {
	@Autowired
	PredictorBean predictorBean;
	
	@Test
	public void simpleTest() throws InterruptedException, ExecutionException, InvalidProtocolBufferException
	{
		PredictorSpec.Builder predictorSpecBuilder = PredictorSpec.newBuilder();
		// PredictorSpecBuilder.setEnabled(true);
		
		predictorSpecBuilder.setName("p1");
		// PredictorSpecBuilder.setRoot("3");
		predictorSpecBuilder.setReplicas(1);
		predictorSpecBuilder.setComponentSpec(PodTemplateSpec.newBuilder());

		PredictiveUnit.Builder predictiveUnitBuilder = PredictiveUnit.newBuilder();
		predictiveUnitBuilder.setName("1");
		predictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		predictiveUnitBuilder.setImplementation(PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
		PredictiveUnit pu1 = predictiveUnitBuilder.build();
		
		predictiveUnitBuilder = PredictiveUnit.newBuilder();
		predictiveUnitBuilder.setName("2");
		predictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		predictiveUnitBuilder.setImplementation(PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
		PredictiveUnit pu2 = predictiveUnitBuilder.build();
		
		
		predictiveUnitBuilder = PredictiveUnit.newBuilder();
		predictiveUnitBuilder.setName("3");
		predictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.ROUTER);
		predictiveUnitBuilder.setImplementation(PredictiveUnit.PredictiveUnitImplementation.RANDOM_ABTEST);
		predictiveUnitBuilder.addChildren(pu1);
		predictiveUnitBuilder.addChildren(pu2);
		Parameter.Builder pBuilder = Parameter.newBuilder().setName("ratioA").setValue("0.5").setType(ParameterType.FLOAT);
		predictiveUnitBuilder.addParameters(pBuilder.build());
		PredictiveUnit pu3 = predictiveUnitBuilder.build();
		
		
		predictorSpecBuilder.setGraph(pu3);


		PredictorSpec predictor = predictorSpecBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

		
		SeldonMessage p = SeldonMessage.newBuilder().build();
			
        SeldonMessage predictorReturn = predictorBean.predict(p,predictorState);
		
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2),0);        
	
	}
}
