package io.seldon.apife.zk;


import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import com.google.common.collect.HashMultimap;
import com.google.common.collect.Multimap;

@Component
public class ZkConfigHandler implements ZkNodeChangeListener{
	  	private static Logger logger = LoggerFactory.getLogger(ZkConfigHandler.class.getName());
	    private final Multimap<String,ConfigUpdateListener> nodeListeners;
	    private final ZkSubscriptionHandler subHandler;
	    private final static String PREFIX = "/config/";
	    
	    @Autowired
	    public ZkConfigHandler( ZkSubscriptionHandler subHandler) {
	        this.nodeListeners = HashMultimap.create();
	        this.subHandler = subHandler;
	    }

	    public void addSubscriber(String node, ConfigUpdateListener listener){
	        if(!nodeListeners.containsKey(node)){
	            nodeListeners.put(node, listener);
	            subHandler.addSubscription(PREFIX+node, this);
	        } else {
	            // already in the list -- just get current value
	            listener.configUpdated(node, subHandler.getValue(PREFIX + node));
	        }
	    }

	    @Override
	    public void nodeChanged(String node, String value) {
	        String genericNodeName = StringUtils.replace(node, PREFIX, "");
	        if(nodeListeners.get(genericNodeName).isEmpty()){
	            logger.warn("Couldn't find listener to tell about zk node change: " + node + " -> " + value);
	        }
	        for (ConfigUpdateListener list : nodeListeners.get(genericNodeName)){
	            list.configUpdated(genericNodeName, value);
	        }
	    }

	    @Override
	    public void nodeDeleted(String node) {
	        // can't think of a reason to implement this
	    }

}
