import argparse
from keras.callbacks import ModelCheckpoint
import numpy as np
import pandas as pd
import pickle
import random
from scipy.io import arff

from model import model

np.random.seed(2018)
np.random.RandomState(2018)
random.seed(2018)

# default args
DATASET = './data/ECG5000_TEST.arff'
SAVE_PATH = './models/'
MODEL_NAME = 'seq2seq'
DATA_RANGE = [0,2627]

# data preprocessing
STANDARDIZED = False
MINMAX = False
CLIP = [99999]

# architecture
TIMESTEPS = 140 # length of 1 ECG
ENCODER_DIM = [20]
DECODER_DIM = [40]
OUTPUT_ACTIVATION = 'sigmoid'

# training
EPOCHS = 100
BATCH_SIZE = 32
LEARNING_RATE = .005
LOSS = 'mean_squared_error'
DROPOUT = 0.
VALIDATION_SPLIT = 0.2
SAVE = False
PRINT_PROGRESS = False
CONTINUE_TRAINING = False
LOAD_PATH = SAVE_PATH

def train(model,X,args):
    """ Train seq2seq-LSTM model. """
    
    # clip data per feature
    for col,clip in enumerate(args.clip):
        X[:,:,col] = np.clip(X[:,:,col],-clip,clip)
    
    # apply scaling and save data preprocessing method
    axis = (0,1) # scaling per feature over all observations
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
            
    # define inputs
    encoder_input_data = X
    decoder_input_data = X
    decoder_target_data = np.roll(X, -1, axis=1) # offset decoder_input_data by 1 across time axis

    # set training arguments
    if args.print_progress:
        verbose = 1
    else:
        verbose = 0

    kwargs = {}
    kwargs['epochs'] = args.epochs
    kwargs['batch_size'] = args.batch_size
    kwargs['shuffle'] = True
    kwargs['validation_split'] = args.validation_split
    kwargs['verbose'] = verbose

    if args.save: # create callback
        print('\nSave stuff')
        checkpointer = ModelCheckpoint(filepath=args.save_path + args.model_name + '_weights.h5',verbose=0,
                                       save_best_only=True,save_weights_only=True)
        kwargs['callbacks'] = [checkpointer]
        
        # save model architecture
        with open(args.save_path + args.model_name + '.pickle', 'wb') as f:
            pickle.dump([X.shape[1],X.shape[2],args.encoder_dim,
                         args.decoder_dim,args.output_activation],f)
    
    model.fit([encoder_input_data, decoder_input_data], decoder_target_data, **kwargs)

def run(args):
    """ Load data, generate training batch, initiate and train model. """
    
    print('\nLoad dataset')
    data = arff.loadarff(args.dataset)
    data = pd.DataFrame(data[0])
    data.drop(columns='target',inplace=True)
    
    print('\nGenerate training batch')
    args.n_features = 1 # only 1 feature in the ECG dataset
    X = data.values[args.data_range[0]:args.data_range[1],:]
    X = np.reshape(X, (X.shape[0],X.shape[1],args.n_features))
    
    print('\nInitiate outlier detector model')
    s2s, enc, dec = model(args.n_features,encoder_dim=args.encoder_dim,decoder_dim=args.decoder_dim,
                          dropout=args.dropout,learning_rate=args.learning_rate,loss=args.loss,
                          output_activation=args.output_activation)
    
    if args.continue_training:
        print('\nLoad pre-trained model')
        s2s.load_weights(args.load_path + args.model_name + '_weights.h5') # load pretrained model weights
    
    if args.print_progress:
        s2s.summary()
    
    print('\nTrain outlier detector')
    train(s2s,X,args)
    
if __name__ == '__main__':
    
    parser = argparse.ArgumentParser(description="Train seq2seq-LSTM outlier detector.")
    parser.add_argument('--dataset',type=str,choices=DATASET,default=DATASET)
    parser.add_argument('--data_range',type=int,nargs=2,default=DATA_RANGE)
    parser.add_argument('--timesteps',type=int,default=TIMESTEPS)
    parser.add_argument('--encoder_dim',type=int,nargs='+',default=ENCODER_DIM)
    parser.add_argument('--decoder_dim',type=int,nargs='+',default=DECODER_DIM)
    parser.add_argument('--output_activation',type=str,default=OUTPUT_ACTIVATION)
    parser.add_argument('--dropout',type=float,default=DROPOUT)
    parser.add_argument('--learning_rate',type=float,default=LEARNING_RATE)
    parser.add_argument('--loss',type=str,default=LOSS)
    parser.add_argument('--validation_split',type=float,default=VALIDATION_SPLIT)
    parser.add_argument('--epochs',type=int,default=EPOCHS)
    parser.add_argument('--batch_size',type=int,default=BATCH_SIZE)
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