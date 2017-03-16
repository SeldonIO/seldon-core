package io.seldon.engine;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import io.seldon.engine.predictors.PredictorBean;
import io.seldon.engine.predictors.PredictorState;
import io.seldon.engine.predictors.PredictorsStore;

@Component
public class AppStartupHandler implements CommandLineRunner{
	
	private PredictorsStore predictorsStore;
	
	@Autowired
	public AppStartupHandler(PredictorsStore predictorsStore){
		this.predictorsStore = predictorsStore;
	}

	@Override
	public void run(String... strings) throws Exception {
		this.createTestSetup();
	}
	
	private void createTestSetup() {
		PredictorState predictor = predictorsStore.retrievePredictorState("digit_classifier_dev");
	}
}
