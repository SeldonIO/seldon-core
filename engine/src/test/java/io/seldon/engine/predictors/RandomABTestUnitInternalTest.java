package io.seldon.engine.predictors;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.SeldonMessage;

public class RandomABTestUnitInternalTest {

	@Test
	public void simpleCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException {
		
		SeldonMessage request = SeldonMessage.newBuilder().build();
		
		PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    	Map<String,PredictiveUnitParameterInterface> params = new HashMap<>();
    	params.put("ratioA", ratioParam);
		
		PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null,new ArrayList<PredictiveUnitState>(), params,null,null, null);
		
		PredictiveUnitState childA = new PredictiveUnitState("A",null,null,null, null, null, null, null);
		PredictiveUnitState childB = new PredictiveUnitState("B",null,null,null, null, null, null, null);
		
		state.addChild(childA);
		state.addChild(childB);

		
		RandomABTestUnit randomABTestUnit = new RandomABTestUnit();
		Method method = RandomABTestUnit.class.getDeclaredMethod("route", SeldonMessage.class, PredictiveUnitState.class);
		method.setAccessible(true);

		// The following values are from random seed 1337
		int routing1 = (int) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals(1,routing1); 
		
		int routing2 = (int) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals(0,routing2);
		
		int routing3 = (int) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals(1,routing3);
	}
	
	@Test(expected=APIException.class)
	public void failureOneChild() throws Throwable{
		
		SeldonMessage request = SeldonMessage.newBuilder().build();
		
		PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    	Map<String,PredictiveUnitParameterInterface> params = new HashMap<>();
    	params.put("ratioA", ratioParam);
		
		PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null,new ArrayList<PredictiveUnitState>(), params, null, null, null);
		
		PredictiveUnitState childA = new PredictiveUnitState("A",null,null,null, null, null, null, null);
		
		state.addChild(childA);
		
		RandomABTestUnit randomABTestUnit = new RandomABTestUnit();
		Method method = RandomABTestUnit.class.getDeclaredMethod("route", SeldonMessage.class, PredictiveUnitState.class);
		method.setAccessible(true);

		// The following should return an error
		try{
			method.invoke(randomABTestUnit, request, state);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}
}
