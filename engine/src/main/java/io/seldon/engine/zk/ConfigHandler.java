package io.seldon.engine.zk;

public interface ConfigHandler {
	void addSubscriber(String node, ConfigUpdateListener listener);
}
