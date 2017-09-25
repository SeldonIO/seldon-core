package io.seldon.engine.metrics;

import java.util.Arrays;

import javax.servlet.http.HttpServletRequest;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpRequest;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.util.StringUtils;

import io.micrometer.core.instrument.Tag;
import io.micrometer.spring.web.client.RestTemplateExchangeTags;
import io.micrometer.spring.web.client.RestTemplateExchangeTagsProvider;
import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.service.InternalPredictionService;

public class SeldonRestTemplateExchangeTagsProvider implements RestTemplateExchangeTagsProvider {

	private final static String PROJECT_ANNOTATION_KEY = "project_name";
	private final static String PREDICTOR_NAME_METRIC = "predictor_name";
	private final static String PREDICTOR_VERSION_METRIC = "predictor_version";
	private final static String MODEL_NAME_METRIC = "model_name";
	private final static String MODEL_IMAGE_METRIC = "model_image";
	private final static String MODEL_VERSION_METRIC = "model_version";
	
	@Autowired
	EnginePredictor enginePredictor;
	
	@Override
	public Iterable<Tag> getTags(String urlTemplate, HttpRequest request, ClientHttpResponse response) 
	{
		Tag uriTag = StringUtils.hasText(urlTemplate)? RestTemplateExchangeTags.uri(urlTemplate): RestTemplateExchangeTags.uri(request);
		
		
	            
		return Arrays.asList(RestTemplateExchangeTags.method(request), uriTag,
				RestTemplateExchangeTags.status(response),
	            RestTemplateExchangeTags.clientName(request),
	            modelName(request),
	            modelImage(request),
	            modelVersion(request),
	            projectName(),
	            predictorName(),
	            predictorVersion());
	}
	
	 public Tag projectName()
	 {
		 return Tag.of("project_name",enginePredictor.getPredictorDef().getAnnotationsOrDefault(PROJECT_ANNOTATION_KEY, "unknown"));
	 }
	
	
	private Tag predictorName()
	{
		if (!StringUtils.hasText(enginePredictor.getPredictorDef().getName()))
			return Tag.of(PREDICTOR_NAME_METRIC, "unknown");
		else
			return Tag.of(PREDICTOR_NAME_METRIC,enginePredictor.getPredictorDef().getName()); 
	}
	
	private Tag predictorVersion()
	{
		if (!StringUtils.hasText(enginePredictor.getPredictorDef().getVersion()))
			return Tag.of(PREDICTOR_VERSION_METRIC, "unknown");
		else
			return Tag.of(PREDICTOR_VERSION_METRIC, enginePredictor.getPredictorDef().getVersion());
	}

	private Tag modelName(HttpRequest request)
	{
		String modelImage = request.getHeaders().getFirst(InternalPredictionService.MODEL_NAME_HEADER);
		if (!StringUtils.hasText(modelImage))
			modelImage = "unknown";
		
		return Tag.of(MODEL_NAME_METRIC, modelImage);
	}
	
	private Tag modelImage(HttpRequest request)
	{
		String modelImage = request.getHeaders().getFirst(InternalPredictionService.MODEL_IMAGE_HEADER);
		if (!StringUtils.hasText(modelImage))
			modelImage = "unknown";
		
		return Tag.of(MODEL_IMAGE_METRIC, modelImage);
	}

	private Tag modelVersion(HttpRequest request)
	{
		String modelVersion = request.getHeaders().getFirst(InternalPredictionService.MODEL_VERSION_HEADER);
		if (!StringUtils.hasText(modelVersion))
			modelVersion = "latest";
		
		return Tag.of(MODEL_VERSION_METRIC, modelVersion);
	}

}
