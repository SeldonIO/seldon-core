# Sequence-to-Sequence LSTM (seq2seq-LSTM) Outlier Algorithm Documentation

The aim of this document is to explain the seq2seq-LSTM algorithm in Seldon's outlier detection framework.

First, we provide a high level overview of the algorithm and the use case, then we will give a detailed explanation of the implementation.

## Overview

Outlier detection has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. The available data is typically unlabeled and detection needs to be done in real-time. The outlier detector can be used as a standalone algorithm, or to detect anomalies in the input data of another predictive model.

The seq2seq-LSTM outlier detection algorithm is suitable for time series data and predicts whether a sequence of input features is an outlier or not, dependent on a threshold level set by the user. The algorithm needs to be pretrained first on a batch of -preferably- inliers.

As observations arrive, the algorithm will:
- clip and scale the input features
- first encode, and then sequentially decode the input time series data in an attempt to reconstruct the initial observations
- compute a reconstruction error between the output of the decoder and the input data
- predict that the observation is an outlier if the error is larger than the threshold level

## Why Sequence-to-Sequence Models?

Seq2seq models convert sequences from one domain into sequences in another domain. A typical example would be sentence translation between different languages. A seq2seq model consists of 2 main building blocks: an encoder and a decoder. The encoder processes the input sequence and initializes the decoder. The decoder then makes sequential predictions for the output sequence. In our case, the decoder aims to reconstruct the input sequence. Both the encoder and decoder are typically implemented with recurrent or 1D convolutional neural networks. Our implementation uses a type of recurrent neural network called LSTM networks. An excellent explanation of how LSTM units work is available [here](http://colah.github.io/posts/2015-08-Understanding-LSTMs/). The loss function to be minimized with stochastic gradient descent is the mean squared error between the input and output sequence, and is called the reconstruction error.

If we train the seq2seq model with inliers, it will be able to replicate new inlier data well with a low reconstruction error. However, if outliers are fed to the seq2seq model, the reconstruction error becomes large and we can classify the sequence as an anomaly.

## Implementation

The implementation is inspired by [this blog post](https://blog.keras.io/a-ten-minute-introduction-to-sequence-to-sequence-learning-in-keras.html).

### 1. Building the seq2seq-LSTM Model

The seq2seq model definition in ```model.py``` takes 4 arguments that define the architecture:
- the number of features in the input
- a list with the number of units per [bidirectional](https://en.wikipedia.org/wiki/Bidirectional_recurrent_neural_networks) LSTM layer in the encoder
- a list with the number of units per LSTM layer in the decoder
- the output activation type for the dense output layer on top of the last LSTM unit in the decoder

``` python
def model(n_features, encoder_dim = [20], decoder_dim = [20], dropout=0., learning_rate=.001, 
          loss='mean_squared_error', output_activation='sigmoid'):
    """ Build seq2seq model.
    
    Arguments:
        - n_features (int): number of features in the data
        - encoder_dim (list): list with number of units per encoder layer
        - decoder_dim (list): list with number of units per decoder layer
        - dropout (float): dropout for LSTM units
        - learning_rate (float): learning rate used during training
        - loss (str): loss function used
        - output_activation (str): activation type for the dense output layer in the decoder
    """
```

First, we define the bidirectional LSTM layers in the encoder and keep the state of the last LSTM unit to initialise the decoder:

```python
# add encoder hidden layers
encoder_lstm = []
for i in range(enc_dim-1):
    encoder_lstm.append(Bidirectional(LSTM(encoder_dim[i], dropout=dropout, 
                                           return_sequences=True,name='encoder_lstm_' + str(i))))
    encoder_hidden = encoder_lstm[i](encoder_hidden)

encoder_lstm.append(Bidirectional(LSTM(encoder_dim[-1], dropout=dropout, return_state=True, 
                                       name='encoder_lstm_' + str(enc_dim-1))))
encoder_outputs, forward_h, forward_c, backward_h, backward_c = encoder_lstm[-1](encoder_hidden)

# only need to keep encoder states
state_h = Concatenate()([forward_h, backward_h])
state_c = Concatenate()([forward_c, backward_c])
encoder_states = [state_h, state_c]
```

We can then define the LSTM units in the decoder, with the states initialised by the encoder:

```python
# initialise decoder states with encoder states
decoder_lstm = []
for i in range(dec_dim):
    decoder_lstm.append(LSTM(decoder_dim[i], dropout=dropout, return_sequences=True,
                             return_state=True, name='decoder_lstm_' + str(i)))
    decoder_hidden, _, _ = decoder_lstm[i](decoder_hidden, initial_state=encoder_states)
```

We add a dense layer with output activation of choice on top of the last LSTM layer in the decoder and compile the model:

```python
# add linear layer on top of LSTM
decoder_dense = Dense(n_features, activation=output_activation, name='dense_output')
decoder_outputs = decoder_dense(decoder_hidden)

# define seq2seq model
model = Model([encoder_inputs, decoder_inputs], decoder_outputs)
optimizer = Adam(lr=learning_rate)
model.compile(optimizer=optimizer, loss=loss)
```

The decoder predictions are sequential and we only need the encoder states to initialise the decoder for the first item in the sequence. From then on, the output and state of the decoder at each step in the sequence is used to predict the next item. As a result, we define separate encoder and decoder models for the prediction stage:

```python
# define encoder model returning encoder states
encoder_model = Model(encoder_inputs, encoder_states * dec_dim)

# define decoder model
# need state inputs for each LSTM layer
decoder_states_inputs = []
for i in range(dec_dim):
    decoder_state_input_h = Input(shape=(decoder_dim[i],), name='decoder_state_input_h_' + str(i))
    decoder_state_input_c = Input(shape=(decoder_dim[i],), name='decoder_state_input_c_' + str(i))
    decoder_states_inputs.append([decoder_state_input_h, decoder_state_input_c])
decoder_states_inputs = [state for states in decoder_states_inputs for state in states]

decoder_inference = decoder_inputs
decoder_states = []
for i in range(dec_dim):
    decoder_inference, state_h, state_c = decoder_lstm[i](decoder_inference, 
                                                          initial_state=decoder_states_inputs[2*i:2*i+2])
    decoder_states.append([state_h,state_c])
decoder_states = [state for states in decoder_states for state in states]

decoder_outputs = decoder_dense(decoder_inference)
decoder_model = Model([decoder_inputs] + decoder_states_inputs,
                      [decoder_outputs] + decoder_states)
```

### 2. Training the model

The seq2seq-LSTM model can be trained on a batch of -ideally- inliers by running the ```train.py``` script with the desired hyperparameters. The example below trains the model on the first 2628 ECG's of the ECG5000 dataset. The input/output sequence has a length of 140, the encoder has 1 bidirectional LSTM layer with 20 units, and the decoder consists of 1 LSTM layer with 40 units. This has to be 2x the number of units of the bidirectional encoder because both the forward and backward encoder states are used to initialise the decoder. Feature-wise minmax scaling between 0 and 1 is applied to the input sequence so we can use a sigmoid activation in the decoder's output layer.

```python
!python train.py \
--dataset './data/ECG5000_TEST.arff' \
--data_range 0 2627 \
--minmax \
--timesteps 140 \
--encoder_dim 20 \
--decoder_dim 40 \
--output_activation 'sigmoid' \
--dropout 0 \
--learning_rate 0.005 \
--loss 'mean_squared_error' \
--epochs 100 \
--batch_size 32 \
--validation_split 0.2 \
--model_name 'seq2seq' \
--print_progress \
--save \
--save_path './models/'
```

The model weights and hyperparameters are saved in the folder specified by "save_path".

### 3. Making predictions

In order to make predictions, which can then be served by Seldon Core, the pre-trained model weights and hyperparameters are loaded when defining an OutlierSeq2SeqLSTM object. The "threshold" argument defines above which reconstruction error a sample is classified as an outlier. The threshold is a key hyperparameter and needs to be picked carefully for each application. The OutlierSeq2SeqLSTM class inherits from the CoreSeq2SeqLSTM class in ```CoreSeq2SeqLSTM.py```.

```python
class CoreSeq2SeqLSTM(object):
    """ Outlier detection using a sequence-to-sequence (seq2seq) LSTM model.
    
    Parameters
    ----------
        threshold (float):  reconstruction error (mse) threshold used to classify outliers
        reservoir_size (int) : number of observations kept in memory using reservoir sampling
     
    Functions
    ----------
        reservoir_sampling : applies reservoir sampling to incoming data
        predict : detect and return outliers
        transform_input : detect outliers and return input features
        send_feedback : add target labels as part of the feedback loop
        tags : add metadata for input transformer
        metrics : return custom metrics
    """
    
    def __init__(self,threshold=0.003,reservoir_size=50000,model_name='seq2seq',load_path='./models/'):
        
        logger.info("Initializing model")
        self.threshold = threshold
        self.reservoir_size = reservoir_size
        self.batch = []
        self.N = 0 # total sample count up until now for reservoir sampling
        self.nb_outliers = 0
        
        # load model architecture parameters
        with open(load_path + model_name + '.pickle', 'rb') as f:
            self.timesteps, self.n_features, encoder_dim, decoder_dim, output_activation = pickle.load(f)
            
        # instantiate model
        self.s2s, self.enc, self.dec = model(self.n_features,encoder_dim=encoder_dim,
                                             decoder_dim=decoder_dim,output_activation=output_activation)
        self.s2s.load_weights(load_path + model_name + '_weights.h5') # load pretrained model weights
        self.s2s._make_predict_function()
        self.enc._make_predict_function()
        self.dec._make_predict_function()
        
        # load data preprocessing info
        with open(load_path + 'preprocess_' + model_name + '.pickle', 'rb') as f:
            preprocess = pickle.load(f)
        self.preprocess, self.clip, self.axis = preprocess[:3]
        if self.preprocess=='minmax':
            self.xmin, self.xmax = preprocess[3:5]
            self.min, self.max = preprocess[5:]
        elif self.preprocess=='standardized':
            self.mu, self.sigma = preprocess[3:]
```

```python
class OutlierSeq2SeqLSTM(CoreSeq2SeqLSTM):
    """ Outlier detection using a sequence-to-sequence (seq2seq) LSTM model.
    
    Parameters
    ----------
        threshold (float) :  reconstruction error (mse) threshold used to classify outliers
        reservoir_size (int) : number of observations kept in memory using reservoir sampling
     
    Functions
    ----------
        send_feedback : add target labels as part of the feedback loop
        metrics : return custom metrics
    """
    def __init__(self,threshold=0.003,reservoir_size=50000,model_name='seq2seq',load_path='./models/'):
        
        super().__init__(threshold=threshold,reservoir_size=reservoir_size,
                         model_name=model_name,load_path=load_path)
```

The actual outlier detection is done by the ```_get_preds``` method which is invoked by ```predict``` or ```transform_input``` dependent on whether the detector is defined as respectively a model or a transformer.

```python
def predict(self, X, feature_names):
        """ Return outlier predictions.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        logger.info("Using component as a model")
        return self._get_preds(X)
```

```python
def transform_input(self, X, feature_names):
    """ Transform the input. 
    Used when the outlier detector sits on top of another model.

    Parameters
    ----------
    X : array-like
    feature_names : array of feature names (optional)
    """
    logger.info("Using component as an outlier-detector transformer")
    self.prediction_meta = self._get_preds(X)
    return X
```

First the data is (optionally) clipped. If the number of observations fed to the outlier detector up until now is at least equal to the defined reservoir size, the feature-wise scaling parameters are updated using the observations in the reservoir. The reservoir is updated each observation using reservoir sampling. We can then scale the input data.

```python
# clip data per feature
for col,clip in enumerate(self.clip):
    X[:,:,col] = np.clip(X[:,:,col],-clip,clip)

# update reservoir
if self.N < self.reservoir_size:
    update_stand = False
else:
    update_stand = True

self.reservoir_sampling(X,update_stand=update_stand)

# apply scaling
if self.preprocess=='minmax':
    X = ((X - self.xmin) / (self.xmax - self.xmin)) * (self.max - self.min) + self.min
elif self.preprocess=='standardized':
    X = (X - self.mu) / (self.sigma + 1e-10)
```

We then make predictions using the ```decode_sequence``` function and calculate the mean squared error between the input and output sequences. If this value is above the threshold, an outlier is predicted.

```python
# make predictions
n_obs = X.shape[0]
self.mse = np.zeros(n_obs)
for obs in range(n_obs):
    input_seq = X[obs:obs+1,:,:]
    decoded_seq = self.decode_sequence(input_seq)
    self.mse[obs] = np.mean(np.power(input_seq[0,:,:] - decoded_seq[0,:,:], 2))
self.prediction = np.array([1 if e > self.threshold else 0 for e in self.mse]).astype(int)
```

The ```decode_sequence``` function takes an input sequence and uses the encoder model to retrieve the state vectors of the last LSTM layer in the encoder so they can be used to initialise the LSTM layers in the decoder. The feature values of the first step in the input sequence are used to initialise the output sequence. We can then use the decoder model to make sequential predictions for the output sequence. At each step, we use the previous step's output value and state as decoder inputs.

```python
def decode_sequence(self,input_seq):
    """ Feed output of encoder to decoder and make sequential predictions. """

    # use encoder the get state vectors
    states_value = self.enc.predict(input_seq)

    # generate initial target sequence
    target_seq = input_seq[0,0,:].reshape((1,1,self.n_features))

    # sequential prediction of time series
    decoded_seq = np.zeros((1, self.timesteps, self.n_features))
    decoded_seq[0,0,:] = target_seq[0,0,:]
    i = 1
    while i < self.timesteps:

        decoder_output = self.dec.predict([target_seq] + states_value)

        # update the target sequence
        target_seq = np.zeros((1, 1, self.n_features))
        target_seq[0, 0, :] = decoder_output[0]

        # update output
        decoded_seq[0, i, :] = decoder_output[0]

        # update states
        states_value = decoder_output[1:]

        i+=1

    return decoded_seq
```

## References

Francois Chollet. A ten-minute introduction to sequence-to-sequence learning in Keras
- https://blog.keras.io/a-ten-minute-introduction-to-sequence-to-sequence-learning-in-keras.html

Christopher Olah. Understanding LSTM Networks
- http://colah.github.io/posts/2015-08-Understanding-LSTMs/

Ilya Sutskever, Oriol Vinyals and Quoc V. Le. Sequence to Sequence Learning with Neural Networks. 2014
- https://arxiv.org/abs/1409.3215