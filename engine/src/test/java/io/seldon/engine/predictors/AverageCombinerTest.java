package io.seldon.engine.predictors;

import static org.hamcrest.CoreMatchers.is;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.junit.Assert;
import org.junit.Test;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionStatusDef;
import io.seldon.protos.PredictionProtos.Tensor;

public class AverageCombinerTest {
	
	@Test
	public void testSimpleCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{
		List<PredictionResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
						.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		Double[] values3 = {2.2,0.9};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values3)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictionResponseDef average = (PredictionResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getResponse().getFeaturesList().get(0),is(names[0]));
		
		Double[][] expected_values = {{(1.0+1.0+2.2)/3,(1.0+0.5+0.9)/3}};
		Assert.assertThat(average.getResponse().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test
	public void testUniqueValue() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictionResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[] values1 = {1.0};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictionResponseDef average = (PredictionResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getResponse().getFeaturesList().get(0),is(names[0]));

		Double[][] expected_values = {{2.0/2}};
		Assert.assertThat(average.getResponse().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test
	public void testUniqueInput() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictionResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[] values1 = {1.0,5.0,0.3};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictionResponseDef average = (PredictionResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getResponse().getFeaturesList().get(0),is(names[0]));

		Double[][] expected_values = {{1.0,5.0,0.3}};
		Assert.assertThat(average.getResponse().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test
	public void testNoInput() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictionResponseDef> predictorReturns = new ArrayList<>();
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictionResponseDef average = (PredictionResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertEquals(0,average.getResponse().getFeaturesList().size());
		Assert.assertEquals(0,average.getResponse().getTensor().getValuesList().size());
	}
	
	@Test
	public void testNoValues() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<PredictionResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {};
		
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.build()).build());
		
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.build()).build());
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		PredictionResponseDef average = (PredictionResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertEquals(0,average.getResponse().getFeaturesList().size());
		Assert.assertEquals(0,average.getResponse().getTensor().getValuesList().size());
	}

    @Test(expected = APIException.class)
	public void testIncompatibleSizes() throws Throwable{
		List<PredictionResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		Double[] values3 = {2.2,0.9,4.5};
		predictorReturns.add(PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
				.setResponse(DefaultDataDef.newBuilder().addAllFeatures(Arrays.asList(names))
						.setTensor(Tensor.newBuilder()
								.addAllValues(Arrays.asList(values3)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			PredictionResponseDef average = (PredictionResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null);
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
    	
    	Assert.assertEquals(average.getResponse().getFeaturesList().size(),0);
		Assert.assertEquals(average.getResponse().getTensor().getValuesList().size(),0);
    	
	}
    
}
