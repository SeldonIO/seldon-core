package io.seldon.engine.predictors;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;

public class RandomABTestUnitInternalTest {

	@Test
	public void simpleCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException {
		
		PredictionRequestDef request = PredictionRequestDef.newBuilder().build();
		
		PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    	Map<String,PredictiveUnitParameterInterface> params = new HashMap<>();
    	params.put("ratioA", ratioParam);
		
		PredictiveUnitState state = new PredictiveUnitState("1","Cool_name",null,params);
		
		PredictiveUnitState childA = new PredictiveUnitState("2","A",null,null);
		PredictiveUnitState childB = new PredictiveUnitState("3","B",null,null);
		
		state.addChild("0", childA);
		state.addChild("1", childB);
		
		RandomABTestUnit randomABTestUnit = new RandomABTestUnit();
		Method method = RandomABTestUnit.class.getDeclaredMethod("forwardPass", PredictionRequestDef.class, PredictiveUnitState.class);
		method.setAccessible(true);

		// The following values are from random seed 1337
		List<PredictiveUnitState> routing1 = (List<PredictiveUnitState>) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals(1,routing1.size());
		Assert.assertEquals("B",routing1.get(0).name); 
		
		List<PredictiveUnitState> routing2 = (List<PredictiveUnitState>) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals(1,routing2.size());
		Assert.assertEquals("A",routing2.get(0).name);
		
		List<PredictiveUnitState> routing3 = (List<PredictiveUnitState>) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals(1,routing3.size());
		Assert.assertEquals("B",routing3.get(0).name);
	}
	
	@Test(expected=APIException.class)
	public void failureOneChild() throws Throwable{
		
		PredictionRequestDef request = PredictionRequestDef.newBuilder().build();
		
		PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    	Map<String,PredictiveUnitParameterInterface> params = new HashMap<>();
    	params.put("ratioA", ratioParam);
		
		PredictiveUnitState state = new PredictiveUnitState("1","Cool_name",null,params);
		
		PredictiveUnitState childA = new PredictiveUnitState("2","A",null,null);
		
		state.addChild("0", childA);
		
		RandomABTestUnit randomABTestUnit = new RandomABTestUnit();
		Method method = RandomABTestUnit.class.getDeclaredMethod("forwardPass", PredictionRequestDef.class, PredictiveUnitState.class);
		method.setAccessible(true);

		// The following should return an error
		try{
			List<PredictiveUnitState> routing1 = (List<PredictiveUnitState>) method.invoke(randomABTestUnit, request, state);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}
}
