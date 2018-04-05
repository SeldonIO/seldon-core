package io.seldon.example.h2o.model;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Component;

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
import io.seldon.wrapper.api.model.SeldonModelHandler;
import io.seldon.wrapper.utils.H2OUtils;

@Component
@Primary
public class H2OModelHandler implements SeldonModelHandler {
	private static Logger logger = LoggerFactory.getLogger(H2OModelHandler.class.getName());
	EasyPredictModelWrapper model;
	
	public H2OModelHandler() throws IOException {
		MojoReaderBackend reader =
                MojoReaderBackendFactory.createReaderBackend(
                  getClass().getClassLoader().getResourceAsStream(
                     "model.zip"), 
                      MojoReaderBackendFactory.CachingStrategy.MEMORY);
		MojoModel modelMojo = ModelMojoReader.readFrom(reader);
		model = new EasyPredictModelWrapper(modelMojo);
		logger.info("Loaded model");
	}
	
	@Override
	public SeldonMessage predict(SeldonMessage payload) {
		List<RowData> rows = H2OUtils.convertSeldonMessage(payload.getData());
		List<AbstractPrediction> predictions = new ArrayList<>();
		for(RowData row : rows)
		{
			try
			{
				BinomialModelPrediction p = model.predictBinomial(row);
				predictions.add(p);
			} catch (PredictException e) {
				logger.info("Error in prediction ",e);
			}
		}
        DefaultData res = H2OUtils.convertH2OPrediction(predictions, payload.getData());

		return SeldonMessage.newBuilder().setData(res).build();
	}

}
