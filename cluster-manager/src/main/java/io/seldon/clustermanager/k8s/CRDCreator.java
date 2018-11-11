package io.seldon.clustermanager.k8s;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.lang.reflect.Type;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import com.google.gson.reflect.TypeToken;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.ApiResponse;
import io.kubernetes.client.Pair;
import io.kubernetes.client.ProgressRequestBody;
import io.kubernetes.client.ProgressResponseBody;
import io.kubernetes.client.models.V1beta1CustomResourceDefinition;
import io.kubernetes.client.util.Config;

@Component
public class CRDCreator {
	protected static Logger logger = LoggerFactory.getLogger(CRDCreator.class.getName());
	public void createCRD() throws IOException, ApiException
	{
		String jsonStr = readFileFromClasspath("crd.json");
		ApiClient client = Config.defaultClient();
		try {
			createCustomResourceDefinition(client,jsonStr.getBytes(),null);
			logger.info("Created CRD");
		} catch (ApiException e) {
			if (e.getCode() == 409)// CRD Already Exists
			{
				logger.info("CRD already exists - ignoring.");
			}
			else if (e.getCode() == 403)// Forbidden - Maybe CRD exists, but we don't know
			{
				logger.warn("No auth to create CRD. Hopefully, one exists.",e); // Hopefully a cluster-wide CRD has been created for us
			}
			else
			{
				logger.warn("Unexpected error trying to create CRD",e);
				throw e;
			}
		}
	}
	private String readFromInputStream(InputStream inputStream)
			  throws IOException {
			    StringBuilder resultStringBuilder = new StringBuilder();
			    try (BufferedReader br
			      = new BufferedReader(new InputStreamReader(inputStream))) {
			        String line;
			        while ((line = br.readLine()) != null) {
			            resultStringBuilder.append(line).append("\n");
			        }
			    }
			  return resultStringBuilder.toString();
			}
	private String readFileFromClasspath(String name) throws IOException
	{
		InputStream in = this.getClass().getClassLoader().getResourceAsStream(name);
		String data = readFromInputStream(in);
		return data;
	}
	
	private String readFile(String path, Charset encoding) 
			  throws IOException 
	 {
		 byte[] encoded = Files.readAllBytes(Paths.get(path));
		 return new String(encoded, encoding);
	 }	

	protected V1beta1CustomResourceDefinition createCustomResourceDefinition(ApiClient apiClient,byte[] body, String pretty)
			throws ApiException {
		ApiResponse<V1beta1CustomResourceDefinition> resp = createCustomResourceDefinitionWithHttpInfo(apiClient,body, pretty);
		return resp.getData();
	}

	private ApiResponse<V1beta1CustomResourceDefinition> createCustomResourceDefinitionWithHttpInfo(ApiClient apiClient,byte[] body,
			String pretty) throws ApiException {
		com.squareup.okhttp.Call call = createCustomResourceDefinitionCall(apiClient,body, pretty, null, null);
		Type localVarReturnType = new TypeToken<V1beta1CustomResourceDefinition>() {
		}.getType();
		return apiClient.execute(call, localVarReturnType);
	}

	public com.squareup.okhttp.Call createCustomResourceDefinitionCall(ApiClient apiClient,byte[] body, String pretty,
			final ProgressResponseBody.ProgressListener progressListener,
			final ProgressRequestBody.ProgressRequestListener progressRequestListener) throws ApiException {
		Object localVarPostBody = body;

		// create path and map variables
		String localVarPath = "/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions";

		List<Pair> localVarQueryParams = new ArrayList<Pair>();
		List<Pair> localVarCollectionQueryParams = new ArrayList<Pair>();
		if (pretty != null)
			localVarQueryParams.addAll(apiClient.parameterToPair("pretty", pretty));

		Map<String, String> localVarHeaderParams = new HashMap<String, String>();

		Map<String, Object> localVarFormParams = new HashMap<String, Object>();

		final String[] localVarAccepts = { "application/json", "application/yaml",
				"application/vnd.kubernetes.protobuf" };
		final String localVarAccept = apiClient.selectHeaderAccept(localVarAccepts);
		if (localVarAccept != null)
			localVarHeaderParams.put("Accept", localVarAccept);

		final String[] localVarContentTypes = { "*/*" };
		final String localVarContentType = apiClient.selectHeaderContentType(localVarContentTypes);
		localVarHeaderParams.put("Content-Type", localVarContentType);

		if (progressListener != null) {
			apiClient.getHttpClient().networkInterceptors().add(new com.squareup.okhttp.Interceptor() {
				@Override
				public com.squareup.okhttp.Response intercept(com.squareup.okhttp.Interceptor.Chain chain)
						throws IOException {
					com.squareup.okhttp.Response originalResponse = chain.proceed(chain.request());
					return originalResponse.newBuilder()
							.body(new ProgressResponseBody(originalResponse.body(), progressListener)).build();
				}
			});
		}

		String[] localVarAuthNames = new String[] { "BearerToken" };
		return apiClient.buildCall(localVarPath, "POST", localVarQueryParams, localVarCollectionQueryParams,
				localVarPostBody, localVarHeaderParams, localVarFormParams, localVarAuthNames, progressRequestListener);
	}

}
