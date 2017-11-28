package io.seldon.engine.predictors;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.Message;

public class RandomABTestUnitInternalTest {

	@Test
	public void simpleCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException {
		
		Message request = Message.newBuilder().build();
		
		PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    	Map<String,PredictiveUnitParameterInterface> params = new HashMap<>();
    	params.put("ratioA", ratioParam);
		
		PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null,new ArrayList<PredictiveUnitState>(), params,null,null);
		
		PredictiveUnitState childA = new PredictiveUnitState("A",null,null,null, null, null, null);
		PredictiveUnitState childB = new PredictiveUnitState("B",null,null,null, null, null, null);
		
		state.addChild(childA);
		state.addChild(childB);
//		
//		Map<String,Integer> emptyDict = new HashMap<String, Integer>();
//		Class<Map<String,Integer>> clazz = (Class<Map<String,Integer>>)(Class)Map.class;

		
		RandomABTestUnit randomABTestUnit = new RandomABTestUnit();
		Method method = RandomABTestUnit.class.getDeclaredMethod("forwardPass", Message.class, PredictiveUnitState.class);
		method.setAccessible(true);

		// The following values are from random seed 1337
		Integer routing1 = (Integer) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals((Integer)1,routing1); 
		
		Integer routing2 = (Integer) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals((Integer)0,routing2);
		
		Integer routing3 = (Integer) method.invoke(randomABTestUnit, request, state);

		Assert.assertEquals((Integer)1,routing3);
	}
	
	@Test(expected=APIException.class)
	public void failureOneChild() throws Throwable{
		
		Message request = Message.newBuilder().build();
		
		PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    	Map<String,PredictiveUnitParameterInterface> params = new HashMap<>();
    	params.put("ratioA", ratioParam);
		
		PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null,new ArrayList<PredictiveUnitState>(), params, null, null);
		
		PredictiveUnitState childA = new PredictiveUnitState("A",null,null,null, null, null, null);
		
		state.addChild(childA);
		
//		Map<String,Integer> emptyDict = new HashMap<String, Integer>();
		
		RandomABTestUnit randomABTestUnit = new RandomABTestUnit();
		Method method = RandomABTestUnit.class.getDeclaredMethod("forwardPass", Message.class, PredictiveUnitState.class);
		method.setAccessible(true);

		// The following should return an error
		try{
			Integer routing1 = (Integer) method.invoke(randomABTestUnit, request, state);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}
}
