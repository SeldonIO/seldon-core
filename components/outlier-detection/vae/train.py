import argparse
from keras.callbacks import ModelCheckpoint
import numpy as np
import pickle
import random

from model import model
from utils import get_kdd_data, generate_batch

np.random.seed(2018)
np.random.RandomState(2018)
random.seed(2018)

# default args
DATASET = 'kddcup99'
SAMPLES = 50000
COLS = str(['srv_count','serror_rate','srv_serror_rate','rerror_rate','srv_rerror_rate',
            'same_srv_rate','diff_srv_rate','srv_diff_host_rate','dst_host_count','dst_host_srv_count',
            'dst_host_same_srv_rate','dst_host_diff_srv_rate','dst_host_same_src_port_rate',
            'dst_host_srv_diff_host_rate','dst_host_serror_rate','dst_host_srv_serror_rate',
            'dst_host_rerror_rate','dst_host_srv_rerror_rate','target'])
MODEL_NAME = 'vae'
SAVE_PATH = './models/'

# data preprocessing
STANDARDIZED = False
MINMAX = False
CLIP = [99999]

# architecture
HIDDEN_LAYERS = 2
LATENT_DIM = 2
HIDDEN_DIM = [15,7]
OUTPUT_ACTIVATION = 'sigmoid'

# training
EPOCHS = 20
BATCH_SIZE = 32
LEARNING_RATE = .001
SAVE = False
PRINT_PROGRESS = False
CONTINUE_TRAINING = False
LOAD_PATH = SAVE_PATH

def train(model,X,args):
    """ Train VAE. """
    
    # clip data per feature
    X = np.clip(X,[-c for c in args.clip],args.clip)
    
    # apply scaling and save data preprocessing method
    axis = 0
    if args.standardized:
        print('\nStandardizing data')
        mu, sigma = np.mean(X,axis=axis), np.std(X,axis=axis)
        X = (X - mu) / (sigma + 1e-10)
        
        with open(args.save_path + 'preprocess_' + args.model_name + '.pickle', 'wb') as f:
            pickle.dump(['standardized',args.clip,axis,mu,sigma], f)
    
    if args.minmax:
        print('\nMinmax scaling of data')
        xmin, xmax = X.min(axis=axis), X.max(axis=axis)
        min, max = 0, 1
        X = ((X - xmin) / (xmax - xmin)) * (max - min) + min
        
        with open(args.save_path + 'preprocess_' + args.model_name + '.pickle', 'wb') as f:
            pickle.dump(['minmax',args.clip,axis,xmin,xmax,min,max], f)

    # set training arguments
    if args.print_progress:
        verbose = 1
    else:
        verbose = 0

    kwargs = {}
    kwargs['epochs'] = args.epochs
    kwargs['batch_size'] = args.batch_size
    kwargs['shuffle'] = True
    kwargs['validation_data'] = (X,None)
    kwargs['verbose'] = verbose

    if args.save: # create callback
        checkpointer = ModelCheckpoint(filepath=args.save_path + args.model_name + '_weights.h5',verbose=0,
                                       save_best_only=True,save_weights_only=True)
        kwargs['callbacks'] = [checkpointer]
            
        # save model architecture
        with open(args.save_path + args.model_name + '.pickle', 'wb') as f:
            pickle.dump([X.shape[1],args.hidden_layers,args.latent_dim,
                         args.hidden_dim,args.output_activation],f)

    model.fit(X,**kwargs)
    
def run(args):
    """ Load data, generate training batch, initiate model and train VAE. """
    
    print('\nLoad dataset')
    if args.dataset=='kddcup99':
        keep_cols = args.keep_cols[1:-1].replace("'","").replace(" ","").split(",")
        data = get_kdd_data(keep_cols=keep_cols)
    else:
        raise ValueError('Only "kddcup99" dataset supported.')
    
    print('\nGenerate training batch')
    X, _ = generate_batch(data,args.samples,0.)
    
    print('\nInitiate outlier detector model')
    n_features = data.shape[1]-1 # nb of features
    vae = model(n_features,hidden_layers=args.hidden_layers,latent_dim=args.latent_dim,hidden_dim=args.hidden_dim,
                output_activation=args.output_activation,learning_rate=args.learning_rate)
    
    if args.continue_training:
        print('\nLoad pre-trained model')
        vae.load_weights(args.load_path + args.model_name + '_weights.h5') # load pretrained model weights
        
    if args.print_progress:
        vae.summary()
    
    print('\nTrain outlier detector')
    train(vae,X,args)
    
if __name__ == '__main__':
    
    parser = argparse.ArgumentParser(description="Train VAE outlier detector.")
    parser.add_argument('--dataset',type=str,choices=DATASET,default=DATASET)
    parser.add_argument('--samples',type=int,default=SAMPLES)
    parser.add_argument('--keep_cols',type=str,default=COLS)
    parser.add_argument('--hidden_layers',type=int,default=HIDDEN_LAYERS)
    parser.add_argument('--latent_dim',type=int,default=LATENT_DIM)
    parser.add_argument('--hidden_dim',type=int,nargs='+',default=HIDDEN_DIM)
    parser.add_argument('--output_activation',type=str,default=OUTPUT_ACTIVATION)
    parser.add_argument('--epochs',type=int,default=EPOCHS)
    parser.add_argument('--batch_size',type=int,default=BATCH_SIZE)
    parser.add_argument('--learning_rate',type=float,default=LEARNING_RATE)
    parser.add_argument('--clip',type=float,nargs='+',default=CLIP)
    parser.add_argument('--standardized', default=STANDARDIZED, action='store_true')
    parser.add_argument('--minmax', default=MINMAX, action='store_true')
    parser.add_argument('--print_progress', default=PRINT_PROGRESS, action='store_true')
    parser.add_argument('--save', default=SAVE, action='store_true')
    parser.add_argument('--save_path',type=str,default=SAVE_PATH)
    parser.add_argument('--load_path',type=str,default=LOAD_PATH)
    parser.add_argument('--model_name',type=str,default=MODEL_NAME)
    parser.add_argument('--continue_training', default=CONTINUE_TRAINING, action='store_true')
    args = parser.parse_args()

    run(args)