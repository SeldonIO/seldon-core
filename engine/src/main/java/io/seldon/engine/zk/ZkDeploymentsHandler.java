package io.seldon.engine.zk;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;

import org.apache.commons.lang3.StringUtils;
import org.apache.curator.framework.CuratorFramework;
import org.apache.curator.framework.recipes.cache.ChildData;
import org.apache.curator.framework.recipes.cache.TreeCacheEvent;
import org.apache.curator.framework.recipes.cache.TreeCacheListener;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

// @Component TODO REMOVE
public class ZkDeploymentsHandler implements TreeCacheListener, DeploymentsHandler{
	private static Logger logger = LoggerFactory.getLogger(ZkDeploymentsHandler.class.getName());
    private final ZkSubscriptionHandler handler;
    private Map<String,Map<String,String>> deploymentsWithInitialConfig;
    private final Set<DeploymentsListener> listeners;
    private final ArrayList<NewDeploymentListener> newDeploymentListeners;
    private static final String DEPLOYMENT_LIST_LOCATION = "deployments";
    private ObjectMapper jsonMapper = new ObjectMapper();

    private boolean initialized = false;


    @Autowired
    public ZkDeploymentsHandler(ZkSubscriptionHandler handler){
        this.handler = handler;
        this.newDeploymentListeners = new ArrayList<>();
        this.listeners = new HashSet<>();
        this.deploymentsWithInitialConfig = new ConcurrentHashMap<>();
    }

    private boolean isDeploymentPath(String path) {
        // i.e. a path like /deployments/testdeployment is one but /deployments/testdeployment/mf is not
        return StringUtils.countMatches(path,"/") == 2;
    }

    @Override
    public Map<String, String> requestCacheDump(String deployment){
        if(initialized) {
            return handler.getChildrenValues("/" + DEPLOYMENT_LIST_LOCATION + "/" + deployment);
        } else {
            return Collections.emptyMap();
        }
    }

    @Override
    public synchronized void addListener(DeploymentsListener listener) {
        logger.info("Adding deployment config listener, current deployments are " + StringUtils.join(deploymentsWithInitialConfig.keySet(),','));
        listeners.add(listener);
    }

    @Override
    public synchronized void addNewDeploymentListener(NewDeploymentListener listener, boolean notifyExistingClients) {
        newDeploymentListeners.add(listener);
        if(notifyExistingClients){
            for(String deployment : deploymentsWithInitialConfig.keySet())
                listener.deploymentAdded(deployment, deploymentsWithInitialConfig.get(deployment));
        }
    }

    @Override
    public synchronized void childEvent(CuratorFramework deployment, TreeCacheEvent event) throws Exception {
        if(event == null || event.getType() == null) {
            logger.warn("Event received was null somewhere");
            return;
        }

        if(!initialized && event.getType() != TreeCacheEvent.Type.INITIALIZED) {
            logger.debug
                    ("Ignore event as we are not in an initialised state: " + event);
            return;
        }
        if(event.getType()== TreeCacheEvent.Type.NODE_ADDED || event.getType() == TreeCacheEvent.Type.NODE_UPDATED)
            logger.info("Message received from ZK : " + event.toString());
        switch (event.getType()){
            case NODE_ADDED:
                if( event.getData() == null || event.getData().getPath()==null) {
                    logger.warn("Event received was null somewhere");
                    return;
                }
                String path = event.getData().getPath();
                if(isDeploymentPath(path)){
                    String deploymentName = retrieveDeploymentName(path);
                    logger.info("Found new deployment : " + deploymentName);
                    for(DeploymentsListener listener: listeners){
                        byte[] data = event.getData().getData();
                        String dataString = data == null ? "" : new String(data);
                        listener.deploymentAdded(deploymentName,dataString);
                    }

                	break;
                } //purposeful cascade as the below deals with the rest of the cases

            case NODE_UPDATED:
                String location = event.getData().getPath();
                boolean foundAMatch = false;
                String[] deploymentAndNode = location.replace("/" + DEPLOYMENT_LIST_LOCATION + "/", "").split("/",2);
                if(deploymentAndNode !=null && deploymentAndNode.length==1){
                    for(DeploymentsListener listener: listeners){
                        foundAMatch = true;
                        byte[] data = event.getData().getData();
                        String dataString = data == null ? "" : new String(data);
                        listener.deploymentUpdated(deploymentAndNode[0],dataString);
                    }

                } else {
                    logger.warn("Couldn't process message for node : " + location + " data : " + new String(event.getData().getData()));
                }
                if (!foundAMatch)
                    logger.warn("Received message for node " + location +" : " + event.getType() + " but found no interested listeners");

                break;
            case NODE_REMOVED:
                path = event.getData().getPath();
                String[] deploymentAndNode2 = path.replace("/" + DEPLOYMENT_LIST_LOCATION + "/", "").split("/");
                if(deploymentAndNode2 !=null && deploymentAndNode2.length==1){
                    for(DeploymentsListener listener: listeners){
                    	listener.deploymentRemoved(deploymentAndNode2[0]);
                    }
                }
                if(isDeploymentPath(path)){
                    String deploymentName = retrieveDeploymentName(path);
                    deploymentsWithInitialConfig.keySet().remove(deploymentName);
                    logger.warn("Deleted deployment : " + deploymentName+" - presently resources will not be released");
                    //for (NewDeploymentListener listener: newDeploymentListeners)
                    //    listener.deploymentDeleted(deploymentName);
                    //jdofactory.deploymentDeleted(deploymentName); // ensure called last in case other deployment removal listeners need db
                }
                break;
            case INITIALIZED:
                initialized = true;
                logger.info("Finished building '/all_deployments' tree cache. ");
                afterCacheBuilt();

        }
    }

    public static String retrieveDeploymentName(String path) {
        return path.replace("/"+DEPLOYMENT_LIST_LOCATION+"/","").split("/")[0];
    }

    private void afterCacheBuilt() throws Exception {
        // first get the deployments
        Collection<ChildData> deploymentChildrenData = handler.getImmediateChildren("/" + DEPLOYMENT_LIST_LOCATION);
        logger.info("Found " +deploymentChildrenData.size() + " deployments on start up.");
        for(ChildData deploymentChildData : deploymentChildrenData) {
            childEvent(null, new TreeCacheEvent(TreeCacheEvent.Type.NODE_ADDED, deploymentChildData));
            // then the children of deployments
            Collection<ChildData> furtherChildren = handler.getChildren(deploymentChildData.getPath());
            logger.info("Found " +furtherChildren.size() + " children for deployment "+ retrieveDeploymentName(deploymentChildData.getPath())+" on startup");
            for (ChildData child : furtherChildren){
                childEvent(null, new TreeCacheEvent(TreeCacheEvent.Type.NODE_ADDED, child));
            }
        }

    }

    public void contextInitialised() throws Exception {
        handler.addSubscription("/" + DEPLOYMENT_LIST_LOCATION, this);
    }
}
