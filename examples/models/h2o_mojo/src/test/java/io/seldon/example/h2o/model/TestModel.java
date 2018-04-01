package io.seldon.example.h2o.model;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

import org.junit.Ignore;
import org.junit.Test;

import com.google.protobuf.ListValue;
import com.google.protobuf.Value;

import hex.genmodel.ModelMojoReader;
import hex.genmodel.MojoModel;
import hex.genmodel.MojoReaderBackend;
import hex.genmodel.MojoReaderBackendFactory;
import hex.genmodel.easy.EasyPredictModelWrapper;
import hex.genmodel.easy.RowData;
import hex.genmodel.easy.exception.PredictException;
import hex.genmodel.easy.prediction.AbstractPrediction;
import hex.genmodel.easy.prediction.BinomialModelPrediction;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;

public class TestModel {

	private EasyPredictModelWrapper getModel(String filename) throws IOException
	{
		MojoReaderBackend reader =
                MojoReaderBackendFactory.createReaderBackend(
                  getClass().getClassLoader().getResourceAsStream(
                     filename), 
                      MojoReaderBackendFactory.CachingStrategy.MEMORY);
		MojoModel modelMojo = ModelMojoReader.readFrom(reader);
		return new EasyPredictModelWrapper(modelMojo);
	}
	
	@Test @Ignore
	public void testMojo() throws Exception {

		
		EasyPredictModelWrapper model = getModel("GLM_model_python_1522573122740_18.zip");


         RowData row = new RowData();
         row.put("AGE", "68");
         row.put("RACE", "2");
         row.put("DCAPS", "2");
         row.put("VOL", "0");
         row.put("GLEASON", "6");

         BinomialModelPrediction p = model.predictBinomial(row);
         System.out.println("Has penetrated the prostatic capsule (1=yes; 0=no): " + p.label);
         System.out.print("Class probabilities: ");
         for (int i = 0; i < p.classProbabilities.length; i++) {
           if (i > 0) {
             System.out.print(",");
           }
           System.out.print(p.classProbabilities[i]);
         }
         System.out.println("");
       }
	
	private SeldonMessage createSeldonMessageTensor()
	{
		Tensor t = Tensor.newBuilder().addShape(1).addShape(5).addValues(68).addValues(2).addValues(2).addValues(0).addValues(6).build();

		SeldonMessage msg = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(t)
				.addNames("AGE")
				.addNames("RACE")
				.addNames("DCAPS")
				.addNames("VOL")
				.addNames("GLEASON")				
				).build();
		return msg;
	}
	
	private SeldonMessage createSeldonMessageNDArray()
	{
		ListValue row = ListValue.newBuilder()
				.addValues(Value.newBuilder().setNumberValue(68))
				.addValues(Value.newBuilder().setNumberValue(2))
				.addValues(Value.newBuilder().setNumberValue(2))
				.addValues(Value.newBuilder().setNumberValue(0))
				.addValues(Value.newBuilder().setNumberValue(6))
				.build();
		ListValue rows = ListValue.newBuilder().addValues(Value.newBuilder().setListValue(row)).build();
		SeldonMessage msg = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setNdarray(rows)
				.addNames("AGE")
				.addNames("RACE")
				.addNames("DCAPS")
				.addNames("VOL")
				.addNames("GLEASON")				
				).build();
		return msg;
	}
	
	
	
	@Test
	public void testMojoFromTensor() throws IOException, PredictException {
		EasyPredictModelWrapper model = getModel("GLM_model_python_1522573122740_18.zip");
		SeldonMessage msg = createSeldonMessageTensor();
		List<RowData> rows = H2OUtils.convertSeldonMessage(msg.getData());
		List<AbstractPrediction> predictions = new ArrayList<>();
		for(RowData row : rows)
		{
			BinomialModelPrediction p = model.predictBinomial(row);
			predictions.add(p);
		}
        DefaultData res = H2OUtils.convertH2OPrediction(predictions, msg.getData());
        System.out.println(res.toString());
	}
	
	@Test
	public void testMojoFromNDArray() throws IOException, PredictException {
		EasyPredictModelWrapper model = getModel("GLM_model_python_1522573122740_18.zip");
		SeldonMessage msg = createSeldonMessageNDArray();
		List<RowData> rows = H2OUtils.convertSeldonMessage(msg.getData());
		List<AbstractPrediction> predictions = new ArrayList<>();
		for(RowData row : rows)
		{
			BinomialModelPrediction p = model.predictBinomial(row);
			predictions.add(p);
		}
        DefaultData res = H2OUtils.convertH2OPrediction(predictions, msg.getData());
        System.out.println(res.toString());
	}
}