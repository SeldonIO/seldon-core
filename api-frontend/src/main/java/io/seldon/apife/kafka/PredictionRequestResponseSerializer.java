package io.seldon.apife.kafka;

import org.apache.kafka.common.serialization.Serializer;

import io.seldon.protos.PredictionProtos.PredictionRequestResponseDef;

public class PredictionRequestResponseSerializer extends Adapter implements Serializer<PredictionRequestResponseDef> {

	@Override
	public byte[] serialize(final String topic, final PredictionRequestResponseDef data) {
		return data.toByteArray();
	}

}
