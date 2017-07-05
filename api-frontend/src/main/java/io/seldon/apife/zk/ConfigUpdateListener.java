package io.seldon.apife.zk;

public interface ConfigUpdateListener {
	void configUpdated(String configKey, String configValue);
}
