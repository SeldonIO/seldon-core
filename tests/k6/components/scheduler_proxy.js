import grpc from 'k6/net/grpc';
import {check} from 'k6';

const schedulerClient = new grpc.Client();
schedulerClient.load(['../../../apis/'], 'mlops/proxy/proxy.proto');

export function connectScheduler(serverUrl) {
    schedulerClient.connect(serverUrl, {
        plaintext: true
    });
    return schedulerClient
}

export function disconnectScheduler() {
    schedulerClient.close();
}

export function loadModel(modelName, data, awaitReady=true) {
    const response = schedulerClient.invoke('seldon.mlops.proxy.SchedulerProxy/LoadModel', data);
    check(response, {'load model success': (r) => r && r.status === grpc.StatusOK})
}

export function unloadModel(modelName, awaitReady = true) {
    const data = {"model":{"name":modelName}}
    const response = schedulerClient.invoke('seldon.mlops.proxy.SchedulerProxy/UnloadModel', data);
    check(response, {'unload model success': (r) => r && r.status === grpc.StatusOK})
}



