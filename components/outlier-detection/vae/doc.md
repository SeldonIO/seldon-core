# Variational Auto-Encoder Outlier (VAE) Algorithm Documentation

The aim of this document is to explain the Variational Auto-Encoder algorithm in Seldon's outlier detection framework.

First, we provide a high level overview of the algorithm and the use case, then we will give a detailed explanation of the implementation.

## Overview

Outlier detection has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. The available data is typically unlabeled and detection needs to be done in real-time. The outlier detector can be used as a standalone algorithm, or to detect anomalies in the input data of another predictive model.

The VAE outlier detection algorithm predicts whether the input features are an outlier or not, dependent on a threshold level set by the user. The algorithm needs to be pretrained first on a batch of -preferably- inliers.

As observations arrive, the algorithm will:
- scale (standardize or minmax) the input features
- first encode, and then decode the input data in an attempt to reconstruct the initial observations
- compute a reconstruction error between the output of the decoder and the input data
- predict that the observation is an outlier if the error is larger than the threshold level

## Why Variational Auto-Encoders?

An Auto-Encoder is an algorithm that consists of 2 main building blocks: an encoder and a decoder. The encoder tries to find a compressed representation of the input data. The compressed data is then fed into the decoder, which aims to replicate the input data. Both the encoder and decoder are typically implemented with neural networks. The loss function to be minimized with stochastic gradient descent is a distance function between the input data and output of the decoder, and is called the reconstruction error.

If we train the Auto-Encoder with inliers, it will be able to replicate new inlier data well with a low reconstruction error. However, if outliers are fed to the Auto-Encoder, the reconstruction error becomes large and we can classify the observation as an anomaly.

A Variational Auto-Encoder adds constraints to the encoded representations of the input. The encodings are parameters of a probability distribution modeling the data. The decoder can then generate new data by sampling from the learned distribution.

## Implementation

### 1. Building the VAE model

The VAE model definition in ```model.py``` takes 4 arguments that define the architecture:
- the number of features in the input
- the number of hidden layers used in the encoder and decoder
- the dimension of the latent variable
- the dimensions of each hidden layer

``` python
def model(n_features, hidden_layers=1, latent_dim=2, hidden_dim=[], 
          output_activation='sigmoid', learning_rate=0.001):
    """ Build VAE model. 
    
    Arguments:
        - n_features (int): number of features in the data
        - hidden_layers (int): number of hidden layers used in encoder/decoder
        - latent_dim (int): dimension of latent variable
        - hidden_dim (list): list with dimension of each hidden layer
        - output_activation (str): activation type for last dense layer in the decoder
        - learning_rate (float): learning rate used during training
    """
```

First, the input data feeds in the encoder and is compressed by mapping it on the latent space which defines the probability distribution of the encodings:

``` python
    # encoder
    inputs = Input(shape=(n_features,), name='encoder_input')
    # define hidden layers
    enc_hidden = Dense(hidden_dim[0], activation='relu', name='encoder_hidden_0')(inputs)
    i = 1
    while i < hidden_layers:
        enc_hidden = Dense(hidden_dim[i],activation='relu',name='encoder_hidden_'+str(i))(enc_hidden)
        i+=1
    
    z_mean = Dense(latent_dim, name='z_mean')(enc_hidden)
    z_log_var = Dense(latent_dim, name='z_log_var')(enc_hidden)
```

We can then sample data from the latent space.

``` python
def sampling(args):
    """ Reparameterization trick by sampling from an isotropic unit Gaussian.
    
    Arguments:
        - args (tensor): mean and log of variance of Q(z|X)
        
    Returns:
        - z (tensor): sampled latent vector
    """
    z_mean, z_log_var = args
    batch = K.shape(z_mean)[0]
    dim = K.int_shape(z_mean)[1]
    epsilon = K.random_normal(shape=(batch, dim)) # by default, random_normal has mean=0 and std=1.0
    return z_mean + K.exp(0.5 * z_log_var) * epsilon # mean + stdev * eps
```

``` python
    # reparametrization trick to sample z
    z = Lambda(sampling, output_shape=(latent_dim,), name='z')([z_mean, z_log_var])
```

The sampled data passes through the decoder which aims to reconstruct the input.

``` python
    # decoder
    latent_inputs = Input(shape=(latent_dim,), name='z_sampling')
    # define hidden layers
    dec_hidden = Dense(hidden_dim[-1], activation='relu', name='decoder_hidden_0')(latent_inputs)

    i = 2
    while i < hidden_layers+1:
        dec_hidden = Dense(hidden_dim[-i],activation='relu',name='decoder_hidden_'+str(i-1))(dec_hidden)
        i+=1

    outputs = Dense(n_features, activation=output_activation, name='decoder_output')(dec_hidden)
```

The loss function is the sum of the reconstruction error and the KL-divergence. While the reconstruction error quantifies how well we can recreate the input data, the KL-divergence measures how close the latent representation is to the unit Gaussian distribution. This trade-off is important because we want our encodings to parameterize a probability distribution from which we can sample data.

``` python
    # define VAE loss, optimizer and compile model
    reconstruction_loss = mse(inputs, outputs)
    reconstruction_loss *= n_features
    kl_loss = 1 + z_log_var - K.square(z_mean) - K.exp(z_log_var)
    kl_loss = K.sum(kl_loss, axis=-1)
    kl_loss *= -0.5
    vae_loss = K.mean(reconstruction_loss + kl_loss)
    vae.add_loss(vae_loss)
```

### 2. Training the model

The VAE model can be trained on a batch of inliers by running the ```train.py``` script with the desired hyperparameters:

``` python
!python train.py \
--dataset 'kddcup99' \
--samples 50000 \
--keep_cols "$cols_str" \
--hidden_layers 1 \
--latent_dim 2 \
--hidden_dim 9 \
--output_activation 'sigmoid' \
--clip 999999 \
--standardized \
--epochs 10 \
--batch_size 32 \
--learning_rate 0.001 \
--print_progress \
--model_name 'vae' \
--save \
--save_path './models/'
```

The model weights and hyperparameters are saved in the folder specified by "save_path".

### 3. Making predictions

In order to make predictions, which can then be served by Seldon Core, the pre-trained model weights and hyperparameters are loaded when defining an OutlierVAE object. The "threshold" argument defines above which reconstruction error a sample is classified as an outlier. The threshold is a key hyperparameter and needs to be picked carefully for each application. The OutlierVAE class inherits from the CoreVAE class in ```CoreVAE.py```.

```python
class CoreVAE(object):
    """ Outlier detection using variational autoencoders (VAE).
    
    Parameters
    ----------
        threshold (float) :  reconstruction error (mse) threshold used to classify outliers
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
    
    def __init__(self,threshold=10,reservoir_size=50000,model_name='vae',load_path='./models/'):
        
        logger.info("Initializing model")
        self.threshold = threshold
        self.reservoir_size = reservoir_size
        self.batch = []
        self.N = 0 # total sample count up until now for reservoir sampling
        self.nb_outliers = 0
        
        # load model architecture parameters
        with open(load_path + model_name + '.pickle', 'rb') as f:
            n_features, hidden_layers, latent_dim, hidden_dim, output_activation = pickle.load(f)
            
        # instantiate model
        self.vae = model(n_features,hidden_layers=hidden_layers,latent_dim=latent_dim,
                         hidden_dim=hidden_dim,output_activation=output_activation)
        self.vae.load_weights(load_path + model_name + '_weights.h5') # load pretrained model weights
        self.vae._make_predict_function()
        
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

``` python
class OutlierVAE(CoreVAE):
    """ Outlier detection using variational autoencoders (VAE).
    
    Parameters
    ----------
        threshold (float) :  reconstruction error (mse) threshold used to classify outliers
        reservoir_size (int) : number of observations kept in memory using reservoir sampling
     
    Functions
    ----------
        send_feedback : add target labels as part of the feedback loop
        metrics : return custom metrics
    """
    
    def __init__(self,threshold=10,reservoir_size=50000,model_name='vae',load_path='./models/'):
        
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

In ```_get_preds```, the observations are first clipped. If the number of observations fed to the outlier detector up until now is at least equal to the defined reservoir size, the feature-wise scaling parameters are updated using the observations in the reservoir. The reservoir is updated each observation using reservoir sampling. The input data is then scaled using either standardization or minmax scaling.

``` python
        # clip data per feature
        X = np.clip(X,[-c for c in self.clip],self.clip)
    
        if self.N < self.reservoir_size:
            update_stand = False
        else:
            update_stand = True
            
        self.reservoir_sampling(X,update_stand=update_stand)
        
        # apply scaling
        if self.preprocess=='minmax':
            X_scaled = ((X - self.xmin) / (self.xmax - self.xmin)) * (self.max - self.min) + self.min
        elif self.preprocess=='standardized':
            X_scaled = (X - self.mu) / (self.sigma + 1e-10)
```

We then make multiple predictions for an observation by sampling N times from the latent space. The mean squared error between the input data and output of the decoder is averaged across the N samples. If this value is above the threshold, an outlier is predicted.

``` python
        # sample latent variables and calculate reconstruction errors
        N = 10
        mse = np.zeros([X.shape[0],N])
        for i in range(N):
            preds = self.vae.predict(X_scaled)
            mse[:,i] = np.mean(np.power(X_scaled - preds, 2), axis=1)
        self.mse = np.mean(mse, axis=1)
        
        # make prediction
        self.prediction = np.array([1 if e > self.threshold else 0 for e in self.mse]).astype(int)
```

## References

Diederik P. Kingma and Max Welling. Auto-Encoding Variational Bayes. ICLR 2014.
- https://arxiv.org/pdf/1312.6114.pdf

Francois Chollet. Building Autoencoders in Keras.
- https://blog.keras.io/building-autoencoders-in-keras.html