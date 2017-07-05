/*
 * Seldon -- open source prediction engine
 * =======================================
 *
 * Copyright 2011-2015 Seldon Technologies Ltd and Rummble Ltd (http://www.seldon.io/)
 *
 * ********************************************************************************************
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * ********************************************************************************************
 */

package io.seldon.apife.zk;

import org.apache.curator.framework.CuratorFramework;
import org.apache.curator.framework.recipes.cache.*;
import org.apache.curator.utils.EnsurePath;
import org.apache.curator.framework.recipes.cache.TreeCacheListener;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import javax.annotation.PreDestroy;
import java.io.IOException;
import java.util.*;

/**
 * @author firemanphil
 *         Date: 06/10/2014
 *         Time: 16:10
 */
@Component
public class ZkSubscriptionHandler {
    private static Logger logger = LoggerFactory.getLogger(ZkSubscriptionHandler.class.getName());
    @Autowired
    private ZkCuratorHandler curator;

    private Map<String,TreeCache> caches = new HashMap<>();
    private Map<String, NodeCache> nodeCaches = new HashMap<>();

    public void addSubscription(String location, TreeCacheListener listener) throws Exception {
        CuratorFramework client = curator.getCurator();
        EnsurePath ensureMvTestPath = new EnsurePath(location);
        ensureMvTestPath.ensure(client.getZookeeperClient());
        TreeCache cache = new TreeCache(client, location);
        cache.getListenable().addListener(listener);
        cache.start();
        caches.put(location, cache);
        logger.info("Added ZooKeeper subscriber for " + location + " children.");
    }

    public boolean addSubscription(final String node, final ZkNodeChangeListener listener) {
        try {
            CuratorFramework client = curator.getCurator();
            final NodeCache cache = new NodeCache(client, node);
            nodeCaches.put(node, cache);
            cache.start(true);
            EnsurePath ensureMvTestPath = new EnsurePath(node);
            ensureMvTestPath.ensure(client.getZookeeperClient());

            logger.info("Added ZooKeeper subscriber for " + node);
            cache.getListenable().addListener(new NodeCacheListener() {
                @Override
                public void nodeChanged() throws Exception {
                    ChildData currentData = cache.getCurrentData();
                    if (currentData == null) {
                        listener.nodeDeleted(node);
                    } else {
                        String data = new String(currentData.getData());
                        listener.nodeChanged(node, data);
                    }

                }
            });
            ChildData data = cache.getCurrentData();
            if(data!=null && data.getData()!=null)
                listener.nodeChanged(node, new String(data.getData()));

            return true;
        } catch (Exception e){
            logger.error("Couldn't add subscription for "+node, e);
            return false;
        }
    }

    public String getValue(String node) {
        if (nodeCaches.containsKey(node)) {
            return new String(nodeCaches.get(node).getCurrentData().getData());
        } else {
            return null;
        }
    }

    public Map<String, String> getChildrenValues(String node){
        logger.info("Getting children values for node " + node);
        Map<String, String> values = new HashMap<>();

            Collection<ChildData> currentData = getChildren(node);
            for(ChildData data : currentData)
            {
                values.put(StringUtils.replace(data.getPath(), node + "/", ""), new String(data.getData()));
            }

        return values;
    }

    public Collection<ChildData> getChildren(String node) {
        return getChildren(node, null);
    }

    private Collection<ChildData> getChildren(String node, TreeCache cache){
        HashSet<ChildData> toReturn = new HashSet<>();

            if(cache==null){
                cache = findParentCache(node);
                if (cache==null) return Collections.EMPTY_LIST;
            }
            Map<String,ChildData> children = cache.getCurrentChildren(node);
            if (children==null)
                return toReturn;
            for (ChildData child : children.values())
                toReturn.addAll(getChildren(child.getPath(), cache));
            toReturn.addAll(children.values());
            return toReturn;

    }

    private TreeCache findParentCache(String node) {
        if(caches.containsKey(node)) return caches.get(node);
        String[] nodeStruct = node.split("/");
        for (int i = 1; i <= nodeStruct.length; i++){
            StringBuffer toRemove = new StringBuffer();
            for (int j = 0; j < i; j++){
                String removal = nodeStruct[nodeStruct.length-i+j];
                toRemove.append("/");
                toRemove.append(removal);

            }
            String lowerNode = node.replace(toRemove.toString(),"");
            if(caches.containsKey(lowerNode)){
                return caches.get(lowerNode);
            }
        }
        return null;

    }

    public void removeSubscription(final String node){
        NodeCache cache = nodeCaches.get(node);
        if(cache!=null){
            try {
                cache.close();
            } catch (IOException e) {
                logger.warn("Problem when removing zookeeper subscription for ");
            }
        }
    }

    @PreDestroy
    public void shutdown() throws IOException {
        for (TreeCache cache : caches.values()){
            cache.close();
        }
        for (NodeCache cache : nodeCaches.values()){
            cache.close();
        }
    }


    public Collection<ChildData> getImmediateChildren(String node) {
        HashSet<ChildData> toReturn = new HashSet<>();

        TreeCache cache = findParentCache(node);
        if (cache==null) return Collections.EMPTY_LIST;

        Map<String,ChildData> children = cache.getCurrentChildren(node);
        if (children!=null)
            toReturn.addAll(children.values());
        return toReturn;
    }
}
