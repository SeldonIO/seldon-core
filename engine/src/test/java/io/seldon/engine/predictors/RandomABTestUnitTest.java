package io.seldon.engine.predictors;

import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamType;
import io.seldon.protos.DeploymentProtos.PredictorDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;

@RunWith(SpringRunner.class)
@SpringBootTest
public class RandomABTestUnitTest {
	@Autowired
	PredictorBean predictorBean;
	
	@Test
	public void simpleTest() throws InterruptedException, ExecutionException
	{
		PredictorDef.Builder predictorDefBuilder = PredictorDef.newBuilder();
		predictorDefBuilder.setEnabled(true);
		
		predictorDefBuilder.setId("p1");
		predictorDefBuilder.setRoot("3");
		predictorDefBuilder.setReplicas(1);

		PredictiveUnitDef.Builder predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();
		predictiveUnitDefBuilder.setId("1");
		predictiveUnitDefBuilder.setType(PredictiveUnitDef.PredictiveUnitType.MODEL);
		predictiveUnitDefBuilder.setSubtype(PredictiveUnitDef.PredictiveUnitSubType.SIMPLE_MODEL);
		PredictiveUnitDef pu1 = predictiveUnitDefBuilder.build();
		
		predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();
		predictiveUnitDefBuilder.setId("2");
		predictiveUnitDefBuilder.setType(PredictiveUnitDef.PredictiveUnitType.MODEL);
		predictiveUnitDefBuilder.setSubtype(PredictiveUnitDef.PredictiveUnitSubType.SIMPLE_MODEL);
		PredictiveUnitDef pu2 = predictiveUnitDefBuilder.build();
		
		
		predictiveUnitDefBuilder = PredictiveUnitDef.newBuilder();
		predictiveUnitDefBuilder.setId("3");
		predictiveUnitDefBuilder.setType(PredictiveUnitDef.PredictiveUnitType.ROUTER);
		predictiveUnitDefBuilder.setSubtype(PredictiveUnitDef.PredictiveUnitSubType.RANDOM_ABTEST);
		predictiveUnitDefBuilder.addChildren("1");
		predictiveUnitDefBuilder.addChildren("2");		
		ParamDef.Builder pBuilder = ParamDef.newBuilder().setName("ratioA").setValue("0.5").setType(ParamType.FLOAT);
		predictiveUnitDefBuilder.addParameters(pBuilder.build());
		PredictiveUnitDef pu3 = predictiveUnitDefBuilder.build();
		
		
		predictorDefBuilder.addPredictiveUnits(pu1);
		predictorDefBuilder.addPredictiveUnits(pu2);
		predictorDefBuilder.addPredictiveUnits(pu3);


		PredictorDef predictor = predictorDefBuilder.build();

		PredictorState predictorState = predictorBean.predictorStateFromDeploymentDef(predictor);

		
		PredictionRequestDef p = PredictionRequestDef.newBuilder().build();
			
        PredictorReturn predictorReturn = predictorBean.predict(p,predictorState);
		
        Assert.assertEquals(SimpleModelUnit.values[0][0], predictorReturn.values[0][0]);
        Assert.assertEquals(SimpleModelUnit.values[0][1], predictorReturn.values[0][1]);
        Assert.assertEquals(SimpleModelUnit.values[0][2], predictorReturn.values[0][2]);        
	
	}
}
