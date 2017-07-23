package io.seldon.engine.predictors;

import static org.hamcrest.CoreMatchers.is;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;

public class AverageCombinerTest {
	
	@Test
	public void testSimpleCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{
		List<PredictorReturn> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[][] values1 = {{1.0,1.0}};
		predictorReturns.add(new PredictorReturn(names,values1));
		
		Double[][] values2 = {{1.0,0.5}};
		predictorReturns.add(new PredictorReturn(names,values2));
		
		Double[][] values3 = {{2.2,0.9}};
		predictorReturns.add(new PredictorReturn(names,values3));
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictorReturn average = (PredictorReturn) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.names,is(names));
		
		Double[][] expected_values = {{(1.0+1.0+2.2)/3,(1.0+0.5+0.9)/3}};
		Assert.assertThat(average.values,is(expected_values));
	}
	
	@Test
	public void testUniqueValue() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictorReturn> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[][] values1 = {{1.0}};
		predictorReturns.add(new PredictorReturn(names,values1));
		
		Double[][] values2 = {{1.0}};
		predictorReturns.add(new PredictorReturn(names,values2));
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictorReturn average = (PredictorReturn) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.names,is(names));
		
		Double[][] expected_values = {{2.0/2}};
		Assert.assertThat(average.values,is(expected_values));
	}
	
	@Test
	public void testUniqueInput() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictorReturn> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[][] values1 = {{1.0,5.0,0.3}};
		predictorReturns.add(new PredictorReturn(names,values1));
		
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictorReturn average = (PredictorReturn) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.names,is(names));
		
		Double[][] expected_values = {{1.0,5.0,0.3}};
		Assert.assertThat(average.values,is(expected_values));
	}
	
	@Test
	public void testNoInput() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictorReturn> predictorReturns = new ArrayList<>();
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictorReturn average = (PredictorReturn) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertNull(average.names);
		Assert.assertNull(average.values);
	}
	
	@Test
	public void testNoValues() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictorReturn> predictorReturns = new ArrayList<>();
		String[] names = {};
		
		Double[][] values1 = {{}};
		predictorReturns.add(new PredictorReturn(names,values1));
		
		Double[][] values2 = {{}};
		predictorReturns.add(new PredictorReturn(names,values2));
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictorReturn average = (PredictorReturn) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.names,is(names));
		Assert.assertThat(average.values,is(values1));
	}

    @Test(expected = APIException.class)
	public void testIncompatibleSizes() throws Throwable{
		List<PredictorReturn> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[][] values1 = {{1.0,1.0}};
		predictorReturns.add(new PredictorReturn(names,values1));
		
		Double[][] values2 = {{1.0,0.5}};
		predictorReturns.add(new PredictorReturn(names,values2));
		
		Double[][] values3 = {{2.2,0.9,4.5}};
		predictorReturns.add(new PredictorReturn(names,values3));
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			PredictorReturn average = (PredictorReturn) method.invoke(averageCombinerUnit, predictorReturns, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
			
	}
	
    @Test
	public void testPredictNoChildren() throws InterruptedException, ExecutionException{
    	
    	PredictionRequestDef p = PredictionRequestDef.newBuilder().build();
    	
    	PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null);
    	
    	AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
    	
    	state.predictiveUnitBean = averageCombinerUnit;

    	Future<PredictionResponseDef> futurePred = averageCombinerUnit.predict(p, state);
    	
    	PredictionResponseDef average = futurePred.get();
    	
    	Assert.assertNull(average.getResponse().getNamesList());
		Assert.assertNull(average.getResponse().getValuesList());
    	
	}
    
}
