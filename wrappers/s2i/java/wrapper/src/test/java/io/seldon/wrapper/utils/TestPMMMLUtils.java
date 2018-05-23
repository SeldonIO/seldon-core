package io.seldon.wrapper.utils;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import javax.xml.bind.JAXBException;

import org.dmg.pmml.PMML;
import org.jpmml.model.PMMLUtil;
import org.junit.Test;
import org.xml.sax.SAXException;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.Tensor;

public class TestPMMMLUtils {

	@Test
	public void testMnist() throws SAXException, JAXBException
	{
		PMML model = PMMLUtil.unmarshal(getClass().getClassLoader().getResourceAsStream(
                "mnist.pmml"));
		
		Tensor.Builder tBuilder = Tensor.newBuilder();
		Random r = new Random();
		List<String> names = new ArrayList<>();
		int numRows = 2;
		for(int rows = 0;rows < numRows;rows++)
			for(int i=0;i<784;i++)
			{
				tBuilder.addValues(r.nextInt() % 255);
				names.add("_c"+i);
			}
		if (numRows == 1)
			tBuilder.addShape(784);
		else
			tBuilder.addShape(numRows).addShape(784);
		DefaultData.Builder dataBuilder = DefaultData.newBuilder();
		dataBuilder.setTensor(tBuilder).addAllNames(names);
		
		PMMLUtils util = new PMMLUtils();
		DefaultData resp = util.evaluate(model, dataBuilder.build());
		System.out.println(resp);
	}
	
}
