import pandas as pd
import numpy as np

from sklearn.preprocessing import LabelEncoder,MinMaxScaler
from sklearn.ensemble import RandomForestClassifier
from xgboost.sklearn import XGBClassifier
from sklearn.neural_network import MLPClassifier
from sklearn.metrics import confusion_matrix,accuracy_score, precision_recall_curve, precision_score, recall_score,roc_curve,f1_score,roc_auc_score,auc
from sklearn.model_selection import KFold
from sklearn.pipeline import Pipeline
from sklearn.externals import joblib
from keras.utils import to_categorical
from transformer import transformer
import matplotlib
import matplotlib.pyplot as plt
import seaborn as sns
import time
import argparse

def build_dataset(df,features,split=0.7):
    
    X =  df[features].as_matrix()
    y = df.isFraud.as_matrix()

    perm = np.random.permutation(len(X))
    X = X[perm,:]
    y = y[perm]
 
    nb_train = int(split*len(X))
    X_train_val = X[:nb_train,:]
    y_train_val = y[:nb_train]
    X_test = X[nb_train:,:]
    y_test = y[nb_train:]

    print 'X_train_val shape:', X_train_val.shape
    print 'y_train_val shape:', y_train_val.shape
    print 'X_test shape:', X_test.shape
    print 'y_test shape:', y_test.shape
    #benchmark accuracy
    bm = 1-y.sum()/float(len(y))
    print 'Benchmark accuracy (predicting all 0):', bm
    
    return X_train_val,X_test,y_train_val,y_test

def calculate_print_scores(y,preds,proba):
    acc_test = accuracy_score(y,preds)
    auc_score_test = roc_auc_score(y,proba[:,1])
    f1_test = f1_score(y,preds)
    precision = precision_score(y,preds)
    recall = recall_score(y,preds)
    cm_test = confusion_matrix(y,preds)
    bm = 1-y.sum()/float(len(y))
    print 'confusion matrix test:'
    print cm_test
    print 'benchmark accuracy (predicting all 0):'
    print bm
    print 'accuracy:'
    print acc_test
    print 'precision:'
    print precision
    print 'recall:'
    print recall
    print 'f1:'
    print f1_test
    print 'roc auc'
    print auc_score_test

def main():
    #load data
    df = pd.read_csv(args.data_path)
    print 'Raw data fields', df.columns
    print 'Nb os samples', len(df)
    print 'Missing values?', df.isnull().values.any()
    print df.head()

    features = ['type','amount','oldbalanceOrg','newbalanceOrig']
    X_train_val,X_test,y_train_val,y_test = build_dataset(df,features)
    print 'sample X_train:', X_train_val[0]
    print 'sample X_test', X_test[0]
    tf = transformer(categorical=True)
    clf = RandomForestClassifier(n_estimators=50,class_weight='balanced',verbose=1)
    p = Pipeline([('trans', tf), ('clf', clf)])

    p.fit(X_train_val,y_train_val)

    filename_p = 'model_pipeline.sav'
    filename_Xtest = '../../explainers/data/paysim_data/test_data/X_test.npy'
    filename_ytest = '../../explainers/data/paysim_data/test_data/y_test.npy'

    joblib.dump(p, filename_p)
    np.save(filename_Xtest,X_test)
    np.save(filename_ytest,y_test)

    p_loaded = joblib.load('model_pipeline.sav')
    preds_test = p_loaded.predict(X_test)
    proba_test = p_loaded.predict_proba(X_test)

    preds_test.sum()/float(len(preds_test))

    calculate_print_scores(y_test,preds_test,proba_test)
    print clf.feature_importances_
if __name__=='__main__':
    parser = argparse.ArgumentParser(prog="create_paysim_pipeline")
    parser.add_argument('--data-path',type=str,help='path to data file',required=True)
    args = parser.parse_args()
    
    main()
