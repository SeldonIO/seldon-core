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

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.StatusDef;
import io.seldon.protos.PredictionProtos.Tensor;

public class AverageCombinerTest {
	
	@Test
	public void testSimpleTensorCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{
		List<ResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
						.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		Double[] values3 = {2.2,0.9};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values3)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		RequestDef prd = RequestDef.newBuilder().build();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, prd, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));
		
		Double[][] expected_values = {{(1.0+1.0+2.2)/3,(1.0+0.5+0.9)/3}};
		Assert.assertEquals(expected_values[0][0],average.getData().getTensor().getValuesList().get(0),1e-7);
	}
	
	@Test
	public void testSimpleNDArrayCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{
		List<ResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setNdarray(ListValue.newBuilder().addValues(Value.newBuilder().setListValue(ListValue.newBuilder()
								.addValues(Value.newBuilder().setNumberValue(values1[0]))
								.addValues(Value.newBuilder().setNumberValue(values1[1])).build())).build()).build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setNdarray(ListValue.newBuilder().addValues(Value.newBuilder().setListValue(ListValue.newBuilder()
								.addValues(Value.newBuilder().setNumberValue(values2[0]))
								.addValues(Value.newBuilder().setNumberValue(values2[1])).build())).build()).build()).build());
		
		Double[] values3 = {2.2,0.9};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setNdarray(ListValue.newBuilder().addValues(Value.newBuilder().setListValue(ListValue.newBuilder()
								.addValues(Value.newBuilder().setNumberValue(values3[0]))
								.addValues(Value.newBuilder().setNumberValue(values3[1])).build())).build()).build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));
		
		Double[][] expected_values = {{(1.0+1.0+2.2)/3,(1.0+0.5+0.9)/3}};
		Assert.assertEquals(expected_values[0][0],average.getData().getNdarray().getValues(0).getListValue().getValues(0).getNumberValue(),1e-7);
	}
	
	@Test
	public void testUniqueValue() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<ResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[] values1 = {1.0};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(1)
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(1)
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));

		Double[][] expected_values = {{2.0/2}};
		Assert.assertThat(average.getData().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test
	public void testUniqueInput() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<ResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[] values1 = {1.0,5.0,0.3};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(3)
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));

		Double[][] expected_values = {{1.0,5.0,0.3}};
		Assert.assertThat(average.getData().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test(expected = APIException.class)
	public void testNoInput() throws Throwable, NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<ResponseDef> predictorReturns = new ArrayList<>();
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}
	
	@Test(expected = APIException.class)
	public void testNoValues() throws Throwable, NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<ResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {};
		
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.build()).build());
		
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.build()).build());
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}

    @Test(expected = APIException.class)
	public void testIncompatibleSizes() throws Throwable{
		List<ResponseDef> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		Double[] values3 = {2.2,0.9,4.5};
		predictorReturns.add(ResponseDef.newBuilder().setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(3)
								.addAllValues(Arrays.asList(values3)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("backwardPass",List.class,RequestDef.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			ResponseDef average = (ResponseDef) method.invoke(averageCombinerUnit, predictorReturns, null, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
			
	}
	
    @Test(expected = APIException.class)
	public void testPredictNoChildren() throws InterruptedException, ExecutionException{
    	
    	RequestDef p = RequestDef.newBuilder().build();
    	
    	PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null,new ArrayList<PredictiveUnitState>(),null,null,null);
    	
    	AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
    	
    	state.predictiveUnitBean = averageCombinerUnit;

    	Future<ResponseDef> futurePred = averageCombinerUnit.predict(p, state, null);
    	
    	ResponseDef average = futurePred.get();
		
    	
	}
    
}
