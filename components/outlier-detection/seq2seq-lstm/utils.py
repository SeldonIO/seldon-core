import collections
import json
import numpy as np
import pandas as pd
import requests
from scipy.io import arff
from sklearn.metrics import confusion_matrix, accuracy_score, f1_score, precision_score, recall_score, fbeta_score

def ecg_data(dataset='TEST',data_range=None, outlier=[2,3,4,5]):
    """ Return ECG dataset with outlier labels. """
    
    data = arff.loadarff('./data/ECG5000_' + dataset + '.arff')
    data = pd.DataFrame(data[0])
    data['target'] = data['target'].astype(int)
    if data_range is None:
        data_range = [0,data.shape[0]]
    outlier_true = data['target'][data_range[0]:data_range[1]].isin(outlier).astype(int).values
    data.drop(columns='target',inplace=True)
    X = data.values[data_range[0]:data_range[1],:]
    return X, outlier_true

def flatten(x):
    """ Flatten list. """
    if isinstance(x, collections.Iterable):
        return [a for i in x for a in flatten(i)]
    else:
        return [x]
    
def performance(y_true,y_pred,roll_window=100):
    """ Return a confusion matrix and calculate rolling accuracy, precision, recall, F1 and F2 scores. """

    # confusion matrix
    cm = confusion_matrix(y_true,y_pred,labels=[0,1])
    tn, fp, fn, tp = cm.ravel()

    # total scores
    acc_tot = accuracy_score(y_true,y_pred)
    prec_tot = precision_score(y_true,y_pred)
    rec_tot = recall_score(y_true,y_pred)
    f1_tot = f1_score(y_true,y_pred)
    f2_tot = fbeta_score(y_true,y_pred,beta=2)

    # rolling scores
    y_true_roll = y_true[-roll_window:]
    y_pred_roll = y_pred[-roll_window:]
    acc_roll = accuracy_score(y_true_roll,y_pred_roll)
    prec_roll = precision_score(y_true_roll,y_pred_roll)
    rec_roll = recall_score(y_true_roll,y_pred_roll)
    f1_roll = f1_score(y_true_roll,y_pred_roll)
    f2_roll = fbeta_score(y_true_roll,y_pred_roll,beta=2)

    scores = [tn, fp, fn, tp, acc_tot, prec_tot, rec_tot, f1_tot, f2_tot,
              acc_roll, prec_roll, rec_roll, f1_roll, f2_roll]
    
    return scores

def outlier_stats(y_true,y_pred,roll_window=100):
    """ Calculate number and percentage of predicted and labeled outliers. """

    y_pred_roll = np.sum(y_pred[-roll_window:])
    y_true_roll = np.sum(y_true[-roll_window:])
    y_pred_tot = np.sum(y_pred)
    y_true_tot = np.sum(y_true)

    return y_pred_roll, y_true_roll, y_pred_tot, y_true_tot

def get_payload(arr):
    features = ["x{}".format(str(i)) for i in range(arr.size)]
    datadef = {"names":features,"ndarray":arr.tolist()}
    payload = {"meta":{},"data":datadef}
    return payload

def rest_request_ambassador(deploymentName,request,endpoint="localhost:8003"):
    response = requests.post(
                "http://"+endpoint+"/seldon/"+deploymentName+"/api/v0.1/predictions",
                json=request)
    print(response.status_code)
    print(response.text)
    return response.json()

def send_feedback_rest(deploymentName,request,response,reward,truth,endpoint="localhost:8003"):
    feedback = {
        "request": request,
        "response": response,
        "reward": reward,
        "truth": {"data":{"ndarray":truth.tolist()}}
    }
    ret = requests.post(
         "http://"+endpoint+"/seldon/"+deploymentName+"/api/v0.1/feedback",
        json=feedback)
    return