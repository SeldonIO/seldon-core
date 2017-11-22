package io.seldon.apife.kafka;

import org.apache.kafka.common.serialization.Serializer;

import io.seldon.protos.PredictionProtos.RequestResponse;

public class RequestResponseSerializer extends Adapter implements Serializer<RequestResponse> {

	@Override
	public byte[] serialize(final String topic, final RequestResponse data) {
		return data.toByteArray();
	}

}
