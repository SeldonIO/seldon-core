package io.seldon.engine.predictors;

import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.Request;
import io.seldon.protos.PredictionProtos.Response;

@RunWith(SpringRunner.class)
@SpringBootTest
public class SimpleModelUnitTest {

	@Autowired
	PredictorBean predictorBean;
	
	@Test
	public void simpleTest() throws InterruptedException, ExecutionException
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
		PredictiveUnitBuilder.setSubtype(PredictiveUnit.PredictiveUnitSubtype.SIMPLE_MODEL);
		
		PredictorSpecBuilder.setGraph(PredictiveUnitBuilder.build());
		PredictorSpec predictor = PredictorSpecBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

		
		Request p = Request.newBuilder().build();

		
        Response predictorReturn = predictorBean.predict(p,predictorState);
        
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2),0);   
		
        
	}
	
}
