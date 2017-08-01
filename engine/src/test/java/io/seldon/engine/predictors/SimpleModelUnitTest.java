package io.seldon.engine.predictors;

import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;

@RunWith(SpringRunner.class)
@SpringBootTest
public class SimpleModelUnitTest {

	@Autowired
	PredictorBean predictorBean;
	
	@Test
	public void simpleTest() throws InterruptedException, ExecutionException
	{
		PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();
		predictorDefBuilder.setEnabled(true);
		
		predictorDefBuilder.setId("p1");
		predictorDefBuilder.setRoot("1");
		predictorDefBuilder.setReplicas(1);
		
		PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();
		predictiveUnitDefBuilder.setId("1");
		predictiveUnitDefBuilder.setType(PredictiveUnitDef.PredictiveUnitType.MODEL);
		predictiveUnitDefBuilder.setSubtype(PredictiveUnitDef.PredictiveUnitSubType.SIMPLE_MODEL);
		
		predictorDefBuilder.addPredictiveUnits(predictiveUnitDefBuilder.build());
		PredictorDef predictor = predictorDefBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromDeploymentDef(predictor);

		
		PredictionRequestDef p = PredictionRequestDef.newBuilder().build();

		
        PredictionResponseDef predictorReturn = predictorBean.predict(p,predictorState);
        
        Assert.assertEquals((double) SimpleModelUnit.values[0], predictorReturn.getResponse().getTensor().getValues(0),0);
        Assert.assertEquals((double) SimpleModelUnit.values[1], predictorReturn.getResponse().getTensor().getValues(1),0);
        Assert.assertEquals((double) SimpleModelUnit.values[2], predictorReturn.getResponse().getTensor().getValues(2),0);   
		
        
	}
	
}
