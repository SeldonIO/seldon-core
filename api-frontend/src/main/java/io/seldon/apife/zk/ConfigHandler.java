package io.seldon.apife.zk;

public interface ConfigHandler {
	void addSubscriber(String node, ConfigUpdateListener listener);
}
