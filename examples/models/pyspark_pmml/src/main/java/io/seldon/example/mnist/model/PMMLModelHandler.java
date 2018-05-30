package io.seldon.example.mnist.model;

import javax.xml.bind.JAXBException;

import org.dmg.pmml.PMML;
import org.jpmml.model.PMMLUtil;
import org.springframework.stereotype.Component;
import org.xml.sax.SAXException;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.wrapper.api.SeldonPredictionService;
import io.seldon.wrapper.utils.PMMLUtils;

@Component
public class PMMLModelHandler implements SeldonPredictionService {
	
	private final PMML model;
	
	public PMMLModelHandler() throws SAXException, JAXBException {
		model = PMMLUtil.unmarshal(getClass().getClassLoader().getResourceAsStream(
                "model.pmml"));
	}

	@Override
	public SeldonMessage predict(SeldonMessage payload) {
		PMMLUtils pmmlUtils = new PMMLUtils();
		DefaultData res = pmmlUtils.evaluate(model, payload.getData());
		return SeldonMessage.newBuilder().setData(res).build();
	}
}
