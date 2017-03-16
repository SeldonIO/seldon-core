package io.seldon.engine.zk;

public interface ConfigUpdateListener {
	void configUpdated(String configKey, String configValue);
}
