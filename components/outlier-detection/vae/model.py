from keras.layers import Lambda, Input, Dense
from keras.models import Model
from keras.losses import mse
from keras import backend as K
from keras.optimizers import Adam
import numpy as np

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
    
    # set dimensions hidden layers
    if hidden_dim==[]:
        i = 0
        dim = n_features
        while i < hidden_layers:
            hidden_dim.append(int(np.max([dim/2,2])))
            dim/=2
            i+=1
    
    # VAE = encoder + decoder
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
    # reparametrization trick to sample z
    z = Lambda(sampling, output_shape=(latent_dim,), name='z')([z_mean, z_log_var])
    # instantiate encoder model
    encoder = Model(inputs, [z_mean, z_log_var, z], name='encoder')
    
    # decoder
    latent_inputs = Input(shape=(latent_dim,), name='z_sampling')
    # define hidden layers
    dec_hidden = Dense(hidden_dim[-1], activation='relu', name='decoder_hidden_0')(latent_inputs)

    i = 2
    while i < hidden_layers+1:
        dec_hidden = Dense(hidden_dim[-i],activation='relu',name='decoder_hidden_'+str(i-1))(dec_hidden)
        i+=1

    outputs = Dense(n_features, activation=output_activation, name='decoder_output')(dec_hidden)
    # instantiate decoder model
    decoder = Model(latent_inputs, outputs, name='decoder')

    # instantiate VAE model
    outputs = decoder(encoder(inputs)[2])
    vae = Model(inputs, outputs, name='vae')
    
    # define VAE loss, optimizer and compile model
    reconstruction_loss = mse(inputs, outputs)
    reconstruction_loss *= n_features
    kl_loss = 1 + z_log_var - K.square(z_mean) - K.exp(z_log_var)
    kl_loss = K.sum(kl_loss, axis=-1)
    kl_loss *= -0.5
    vae_loss = K.mean(reconstruction_loss + kl_loss)
    vae.add_loss(vae_loss)
    
    optimizer = Adam(lr=learning_rate)
    vae.compile(optimizer=optimizer)
    
    return vae