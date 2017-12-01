package io.seldon.engine.predictors;

import static org.hamcrest.CoreMatchers.is;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.ExecutionException;

import org.junit.Assert;
import org.junit.Test;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Status;
import io.seldon.protos.PredictionProtos.Tensor;

public class AverageCombinerTest {
	
	@Test
	public void testSimpleTensorCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{
		List<SeldonMessage> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
						.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		Double[] values3 = {2.2,0.9};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values3)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		SeldonMessage average = (SeldonMessage) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));
		
		Double[][] expected_values = {{(1.0+1.0+2.2)/3,(1.0+0.5+0.9)/3}};
		Assert.assertEquals(expected_values[0][0],average.getData().getTensor().getValuesList().get(0),1e-7);
	}
	
	@Test
	public void testSimpleNDArrayCase() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{
		List<SeldonMessage> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setNdarray(ListValue.newBuilder().addValues(Value.newBuilder().setListValue(ListValue.newBuilder()
								.addValues(Value.newBuilder().setNumberValue(values1[0]))
								.addValues(Value.newBuilder().setNumberValue(values1[1])).build())).build()).build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setNdarray(ListValue.newBuilder().addValues(Value.newBuilder().setListValue(ListValue.newBuilder()
								.addValues(Value.newBuilder().setNumberValue(values2[0]))
								.addValues(Value.newBuilder().setNumberValue(values2[1])).build())).build()).build()).build());
		
		Double[] values3 = {2.2,0.9};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setNdarray(ListValue.newBuilder().addValues(Value.newBuilder().setListValue(ListValue.newBuilder()
								.addValues(Value.newBuilder().setNumberValue(values3[0]))
								.addValues(Value.newBuilder().setNumberValue(values3[1])).build())).build()).build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		SeldonMessage average = (SeldonMessage) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));
		
		Double[][] expected_values = {{(1.0+1.0+2.2)/3,(1.0+0.5+0.9)/3}};
		Assert.assertEquals(expected_values[0][0],average.getData().getNdarray().getValues(0).getListValue().getValues(0).getNumberValue(),1e-7);
	}
	
	@Test
	public void testUniqueValue() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<SeldonMessage> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[] values1 = {1.0};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(1)
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(1)
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		SeldonMessage average = (SeldonMessage) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));

		Double[][] expected_values = {{2.0/2}};
		Assert.assertThat(average.getData().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test
	public void testUniqueInput() throws NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<SeldonMessage> predictorReturns = new ArrayList<>();
		String[] names = {"c"};
		
		Double[] values1 = {1.0,5.0,0.3};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(3)
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		SeldonMessage average = (SeldonMessage) method.invoke(averageCombinerUnit, predictorReturns, null);
		
		Assert.assertThat(average.getData().getNamesList().get(0),is(names[0]));

		Double[][] expected_values = {{1.0,5.0,0.3}};
		Assert.assertThat(average.getData().getTensor().getValuesList().get(0),is(expected_values[0][0]));
	}
	
	@Test(expected = APIException.class)
	public void testNoInput() throws Throwable, NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<SeldonMessage> predictorReturns = new ArrayList<>();
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			method.invoke(averageCombinerUnit, predictorReturns, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}
	
	@Test(expected = APIException.class)
	public void testNoValues() throws Throwable, NoSuchMethodException, SecurityException, IllegalAccessException, IllegalArgumentException, InvocationTargetException{

		List<SeldonMessage> predictorReturns = new ArrayList<>();
		String[] names = {};
		
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.build()).build());
		
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.build()).build());
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			method.invoke(averageCombinerUnit, predictorReturns, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
	}

    @Test(expected = APIException.class)
	public void testIncompatibleSizes() throws Throwable{
		List<SeldonMessage> predictorReturns = new ArrayList<>();
		String[] names = {"c","d"};
		
		Double[] values1 = {1.0,1.0};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values1)).build())
						.build()).build());
		
		Double[] values2 = {1.0,0.5};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(2)
								.addAllValues(Arrays.asList(values2)).build())
						.build()).build());
		
		Double[] values3 = {2.2,0.9,4.5};
		predictorReturns.add(SeldonMessage.newBuilder().setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
				.setData(DefaultData.newBuilder().addAllNames(Arrays.asList(names))
						.setTensor(Tensor.newBuilder().addShape(1).addShape(3)
								.addAllValues(Arrays.asList(values3)).build())
						.build()).build());
		
		AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
		Method method = AverageCombinerUnit.class.getDeclaredMethod("aggregateOutputs",List.class,PredictiveUnitState.class);
		method.setAccessible(true);
		
		try{
			method.invoke(averageCombinerUnit, predictorReturns, null);
		}
		catch( InvocationTargetException e){
			Throwable targetException = e.getTargetException();
			throw targetException;
		}
			
	}
	
    @Test
	public void testPredictNoChildren() throws InterruptedException, ExecutionException{
    	
    	SeldonMessage p = SeldonMessage.newBuilder().build();
    	
    	PredictiveUnitState state = new PredictiveUnitState("Cool_name",null,null,new ArrayList<PredictiveUnitState>(),null,null,null,null);
    	
    	AverageCombinerUnit averageCombinerUnit = new AverageCombinerUnit();
    	
    	state.predictiveUnitBean = averageCombinerUnit;

    	averageCombinerUnit.getOutput(p, state);
	}
    
}
