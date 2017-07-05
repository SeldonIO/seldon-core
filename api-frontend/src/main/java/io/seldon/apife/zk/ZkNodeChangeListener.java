package io.seldon.apife.zk;

public interface ZkNodeChangeListener {

	   void nodeChanged(String node, String value);

	   void nodeDeleted(String node);
	   
}
