/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.metrics;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Map;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpRequest;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.boot.actuate.metrics.web.client.RestTemplateExchangeTags;
import org.springframework.boot.actuate.metrics.web.client.RestTemplateExchangeTagsProvider;

import io.micrometer.core.instrument.Tag;

import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.predictors.PredictiveUnitState;
import io.seldon.engine.service.InternalPredictionService;

@Component
public class SeldonRestTemplateExchangeTagsProvider implements RestTemplateExchangeTagsProvider {

	private final static String DEPLOYMENT_NAME_METRIC = "deployment_name";
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
	            deploymentName(),
	            modelName(request),
	            modelImage(request),
	            modelVersion(request),
		        predictorName(),
	            predictorVersion());
	}
	
	public Iterable<Tag> getModelMetrics(PredictiveUnitState state, Map<String,String> customTags)
	{
		ArrayList<Tag> customTagsList = new ArrayList<>();
		if (customTags != null)
			for(Map.Entry<String, String> e : customTags.entrySet())
			{
				customTagsList.add(Tag.of(e.getKey(), e.getValue()));
			}
		for(Tag t : getModelMetrics(state))
		{
			customTagsList.add(t);
		}
		return customTagsList;
	}
	
	public Iterable<Tag> getModelMetrics(PredictiveUnitState state)
	{
		return Arrays.asList(
				 deploymentName(),
				 predictorName(),
				 predictorVersion(),
				 modelName(state.name),
				 modelImage(state.imageName),
				 modelVersion(state.imageVersion));
	}
	
	
    private Tag deploymentName()
    {
    	return Tag.of(DEPLOYMENT_NAME_METRIC, enginePredictor.getDeploymentName());
    }
	
	
	private Tag predictorName()
	{
		if (!StringUtils.hasText(enginePredictor.getPredictorSpec().getName()))
			return Tag.of(PREDICTOR_NAME_METRIC, "unknown");
		else
			return Tag.of(PREDICTOR_NAME_METRIC,enginePredictor.getPredictorSpec().getName()); 
	}
	
	private Tag predictorVersion()
	{
		if (!StringUtils.hasText(enginePredictor.getPredictorSpec().getLabelsOrDefault("version", "")))
			return Tag.of(PREDICTOR_VERSION_METRIC, "unknown");
		else
			return Tag.of(PREDICTOR_VERSION_METRIC, enginePredictor.getPredictorSpec().getLabelsOrDefault("version", ""));
	}

	private Tag modelName(HttpRequest request)
	{
		String modelName = request.getHeaders().getFirst(InternalPredictionService.MODEL_NAME_HEADER);
		return modelName(modelName);
	}
	
	private Tag modelName(String modelName)
	{
		if (!StringUtils.hasText(modelName))
			modelName = "unknown";
		return Tag.of(MODEL_NAME_METRIC, modelName);
	}
	
	private Tag modelImage(HttpRequest request)
	{
		String modelImage = request.getHeaders().getFirst(InternalPredictionService.MODEL_IMAGE_HEADER);
		return modelImage(modelImage);
	}
	
	private Tag modelImage(String modelImage)
	{
		if (!StringUtils.hasText(modelImage))
			modelImage = "unknown";
		
		return Tag.of(MODEL_IMAGE_METRIC, modelImage);
	}

	private Tag modelVersion(HttpRequest request)
	{
		String modelVersion = request.getHeaders().getFirst(InternalPredictionService.MODEL_VERSION_HEADER);
		return modelVersion(modelVersion);
	}
	
	private Tag modelVersion(String modelVersion)
	{
		if (!StringUtils.hasText(modelVersion))
			modelVersion = "latest";
		
		return Tag.of(MODEL_VERSION_METRIC, modelVersion);
	}
	
	

}
