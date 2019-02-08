import argparse
import numpy as np
import pickle
import random
from sklearn.ensemble import IsolationForest

from utils import get_kdd_data, generate_batch

np.random.seed(2018)
np.random.RandomState(2018)
random.seed(2018)

# default args
DATASET = 'kddcup99'
SAMPLES = 50000
COLS = str(['duration','protocol_type','flag','src_bytes','dst_bytes','land','wrong_fragment','urgent','hot',
    'num_failed_logins','logged_in','num_compromised','root_shell','su_attempted','num_root','num_file_creations',
    'num_shells','num_access_files','num_outbound_cmds','is_host_login','is_guest_login','count','srv_count',
    'serror_rate','srv_serror_rate','rerror_rate','srv_rerror_rate','same_srv_rate','diff_srv_rate',
    'srv_diff_host_rate','dst_host_count','dst_host_srv_count','dst_host_same_srv_rate','dst_host_diff_srv_rate',
    'dst_host_same_src_port_rate','dst_host_srv_diff_host_rate','dst_host_serror_rate','dst_host_srv_serror_rate',
    'dst_host_rerror_rate','dst_host_srv_rerror_rate','target'])
MODEL_NAME = 'if'
SAVE = True
SAVE_PATH = './models/'

# Isolation Forest hyperparameters
CONTAMINATION = .1
N_ESTIMATORS = 50
MAX_SAMPLES = .8
MAX_FEATURES = 1.

def train(X,args):
    """ Fit Isolation Forest. """
    
    clf = IsolationForest(n_estimators=args.n_estimators, max_samples=args.max_samples, max_features=args.max_features,
                          contamination=args.contamination,behaviour='new')
    clf.fit(X)
    
    if args.save: # save model
        with open(args.save_path + args.model_name + '.pickle', 'wb') as f:
            pickle.dump(clf,f)

def run(args):
    """ Load data, generate training batch and train Isolation Forest. """
    
    print('\nLoad dataset')
    if args.dataset=='kddcup99':
        keep_cols = args.keep_cols[1:-1].replace("'","").replace(" ","").split(",")
        data = get_kdd_data(keep_cols=keep_cols)
    else:
        raise ValueError('Only "kddcup99" dataset supported.')
    
    print('\nGenerate training batch')
    X, _ = generate_batch(data,args.samples,args.contamination)
    
    print('\nTrain outlier detector')
    train(X,args)
    
    print('\nTraining done!')

if __name__ == '__main__':
    
    parser = argparse.ArgumentParser(description="Train Isolation Forest outlier detector.")
    parser.add_argument('--dataset',type=str,choices=DATASET,default=DATASET)
    parser.add_argument('--keep_cols',type=str,default=COLS)
    parser.add_argument('--samples',type=int,default=SAMPLES)
    parser.add_argument('--contamination',type=float,default=CONTAMINATION)
    parser.add_argument('--n_estimators',type=int,default=N_ESTIMATORS)
    parser.add_argument('--max_samples',type=float,default=MAX_SAMPLES)
    parser.add_argument('--max_features',type=float,default=MAX_FEATURES)
    parser.add_argument('--model_name',type=str,default=MODEL_NAME)
    parser.add_argument('--save', default=SAVE, action='store_false')
    parser.add_argument('--save_path',type=str,default=SAVE_PATH)
    args = parser.parse_args()

    run(args)