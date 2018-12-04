import collections
import json
import numpy as np
import pandas as pd
import requests
from sklearn.datasets import fetch_kddcup99
from sklearn.metrics import confusion_matrix, accuracy_score, f1_score, precision_score, recall_score, fbeta_score

pd.options.mode.chained_assignment = None  # default='warn'

def get_kdd_data(target=['dos','r2l','u2r','probe'],
                 keep_cols=['srv_count','serror_rate','srv_serror_rate','rerror_rate','srv_rerror_rate',
                            'same_srv_rate','diff_srv_rate','srv_diff_host_rate','dst_host_count','dst_host_srv_count',
                            'dst_host_same_srv_rate','dst_host_diff_srv_rate','dst_host_same_src_port_rate',
                            'dst_host_srv_diff_host_rate','dst_host_serror_rate','dst_host_srv_serror_rate',
                            'dst_host_rerror_rate','dst_host_srv_rerror_rate','target'],
                percent10=False):
    """ Load KDD Cup 1999 data and return in dataframe. """
    
    data_raw = fetch_kddcup99(subset=None, data_home=None, percent10=percent10)    
    
    # specify columns
    cols=['duration','protocol_type','service','flag','src_bytes','dst_bytes','land','wrong_fragment','urgent','hot',
          'num_failed_logins','logged_in','num_compromised','root_shell','su_attempted','num_root','num_file_creations',
          'num_shells','num_access_files','num_outbound_cmds','is_host_login','is_guest_login','count','srv_count',
          'serror_rate','srv_serror_rate','rerror_rate','srv_rerror_rate','same_srv_rate','diff_srv_rate',
          'srv_diff_host_rate','dst_host_count','dst_host_srv_count','dst_host_same_srv_rate','dst_host_diff_srv_rate',
          'dst_host_same_src_port_rate','dst_host_srv_diff_host_rate','dst_host_serror_rate','dst_host_srv_serror_rate',
          'dst_host_rerror_rate','dst_host_srv_rerror_rate']
    
    # create dataframe
    data = pd.DataFrame(data=data_raw['data'],columns=cols)
    
    # add target to dataframe
    data['attack_type'] = data_raw['target']
    
    # specify and map attack types
    attack_list = np.unique(data['attack_type'])
    attack_category = ['dos','u2r','r2l','r2l','r2l','probe','dos','u2r','r2l','dos','probe','normal','u2r',
                       'r2l','dos','probe','u2r','probe','dos','r2l','dos','r2l','r2l']
    
    attack_types = {}
    for i,j in zip(attack_list,attack_category):
        attack_types[i] = j

    data['attack_category'] = 'normal'
    for key,value in attack_types.items():
        data['attack_category'][data['attack_type'] == key] = value
    
    # define target
    data['target'] = 0
    for t in target:
        data['target'][data['attack_category'] == t] = 1
    
    # define columns to be dropped
    drop_cols = []
    for col in data.columns.values:
        if col not in keep_cols:
            drop_cols.append(col)
    
    if drop_cols!=[]:
        data.drop(columns=drop_cols,inplace=True)
    
    # apply OHE if necessary
    cols_ohe = ['protocol_type','service','flag']
    for col in cols_ohe:
        if col in keep_cols:
            col_ohe = pd.get_dummies(data[col],prefix=col)
            data = data.join(col_ohe)
            data.drop([col],axis=1,inplace=True)
    
    return data


def sample_df(df,n):
    """ Sample from df. """
    if n < df.shape[0]+1:
        replace = False
    else:
        replace = True
    return df.sample(n=n,replace=replace)


def generate_batch(data,n_samples,frac_outliers):
    """ Generate random batch from data with fixed size and fraction of outliers. """
    
    normal = data[data['target']==0]
    outlier = data[data['target']==1]
    
    if n_samples==1:
        n_outlier = np.random.binomial(1,frac_outliers)
        n_normal = 1 - n_outlier
    else:
        n_normal = int((1-frac_outliers) * n_samples)
        n_outlier = int(frac_outliers * n_samples)
    
    batch_normal = sample_df(normal,n_normal)
    batch_outlier = sample_df(outlier,n_outlier)
    
    batch = pd.concat([batch_normal,batch_outlier])
    batch = batch.sample(frac=1).reset_index(drop=True)
    
    outlier_true = batch['target'].values
    batch.drop(columns=['target'],inplace=True)
    
    return batch.values.astype('float'), outlier_true

def flatten(x):
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
    features = ["srv_count","serror_rate","srv_serror_rate","rerror_rate","srv_rerror_rate","same_srv_rate",
             "diff_srv_rate","srv_diff_host_rate","dst_host_count","dst_host_srv_count","dst_host_same_srv_rate",
             "dst_host_diff_srv_rate","dst_host_same_src_port_rate","dst_host_srv_diff_host_rate",
             "dst_host_serror_rate","dst_host_srv_serror_rate","dst_host_rerror_rate","dst_host_srv_rerror_rate"]
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