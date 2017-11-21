package io.seldon.engine.predictors;

import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.protos.DeploymentProtos.Parameter;
import io.seldon.protos.DeploymentProtos.Parameter.ParameterType;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.ResponseDef;

@RunWith(SpringRunner.class)
@SpringBootTest
public class RandomABTestUnitTest {
	@Autowired
	PredictorBean predictorBean;
	
	@Test
	public void simpleTest() throws InterruptedException, ExecutionException
	{
		PredictorSpec.Builder PredictorSpecBuilder = PredictorSpec.newBuilder();
		// PredictorSpecBuilder.setEnabled(true);
		
		PredictorSpecBuilder.setName("p1");
		// PredictorSpecBuilder.setRoot("3");
		PredictorSpecBuilder.setReplicas(1);
		PredictorSpecBuilder.setComponentSpec(PodTemplateSpec.newBuilder());

		PredictiveUnit.Builder PredictiveUnitBuilder = PredictiveUnit.newBuilder();
		PredictiveUnitBuilder.setName("1");
		PredictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		PredictiveUnitBuilder.setSubtype(PredictiveUnit.PredictiveUnitSubtype.SIMPLE_MODEL);
		PredictiveUnit pu1 = PredictiveUnitBuilder.build();
		
		PredictiveUnitBuilder = PredictiveUnit.newBuilder();
		PredictiveUnitBuilder.setName("2");
		PredictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
		PredictiveUnitBuilder.setSubtype(PredictiveUnit.PredictiveUnitSubtype.SIMPLE_MODEL);
		PredictiveUnit pu2 = PredictiveUnitBuilder.build();
		
		
		PredictiveUnitBuilder = PredictiveUnit.newBuilder();
		PredictiveUnitBuilder.setName("3");
		PredictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.ROUTER);
		PredictiveUnitBuilder.setSubtype(PredictiveUnit.PredictiveUnitSubtype.RANDOM_ABTEST);
		PredictiveUnitBuilder.addChildren(pu1);
		PredictiveUnitBuilder.addChildren(pu2);
		Parameter.Builder pBuilder = Parameter.newBuilder().setName("ratioA").setValue("0.5").setType(ParameterType.FLOAT);
		PredictiveUnitBuilder.addParameters(pBuilder.build());
		PredictiveUnit pu3 = PredictiveUnitBuilder.build();
		
		
		PredictorSpecBuilder.setGraph(pu3);


		PredictorSpec predictor = PredictorSpecBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

		
		RequestDef p = RequestDef.newBuilder().build();
			
        ResponseDef predictorReturn = predictorBean.predict(p,predictorState);
		
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2),0);        
	
	}
}
