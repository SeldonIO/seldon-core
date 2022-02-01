import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const schedulerClient = new grpc.Client();
schedulerClient.load(['../../../apis/mlops/scheduler/'], 'scheduler.proto');

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
    //console.log(data)
    const response = schedulerClient.invoke('seldon.mlops.scheduler.Scheduler/LoadModel', data);
    if (check(response, {'load model success': (r) => r && r.status === grpc.StatusOK})) {
        if (awaitReady) {
            awaitStatus(modelName, "ModelAvailable")
        }
    }
}

export function getModelStatus(modelName) {
    const data = {"model":{"name":modelName}}
    const response = schedulerClient.invoke('seldon.mlops.scheduler.Scheduler/ModelStatus', data);
    if (check(response, {'model status success': (r) => r && r.status === grpc.StatusOK})) {
        const responseData = response.message
        //console.log(JSON.stringify(response.message));

        if (responseData.versions.length !== 1) {
            return ""
        } else {
            return responseData.versions[0].state.state
        }
    } else {
        return ""
    }
}

export function awaitStatus(modelName, status) {
    while (getModelStatus(modelName) !== status) {
        sleep(1)
    }
}

export function unloadModel(modelName, awaitReady = true) {
    const data = {"model":{"name":modelName}}
    //console.log(JSON.stringify(data))
    const response = schedulerClient.invoke('seldon.mlops.scheduler.Scheduler/UnloadModel', data);
    if (check(response, {'unload model success': (r) => r && r.status === grpc.StatusOK})) {
        if (awaitReady) {
            awaitStatus(modelName, "ModelTerminated")
        }
    }
}



