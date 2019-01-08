from keras.layers import Input, LSTM, Dense, Bidirectional, Concatenate
from keras.models import Model
from keras.optimizers import Adam
import numpy as np

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
    
    enc_dim = len(encoder_dim)
    dec_dim = len(decoder_dim)
    
    # seq2seq = encoder + decoder
    # encoder
    encoder_hidden = encoder_inputs = Input(shape=(None, n_features), name='encoder_input')
    
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
    
    # decoder
    decoder_hidden = decoder_inputs = Input(shape=(None, n_features), name='decoder_input')
    
    # add decoder hidden layers
    # check if dimensions are correct
    dim_check = [(idx,dim) for idx,dim in enumerate(decoder_dim) if dim!=encoder_dim[-1]*2]
    if len(dim_check)>0:
        raise ValueError('\nDecoder (layer,units) {0} is not compatible with encoder hidden ' \
                         'states. Units should be equal to {1}'.format(dim_check,encoder_dim[-1]*2))
    
    # initialise decoder states with encoder states
    decoder_lstm = []
    for i in range(dec_dim):
        decoder_lstm.append(LSTM(decoder_dim[i], dropout=dropout, return_sequences=True,
                                 return_state=True, name='decoder_lstm_' + str(i)))
        decoder_hidden, _, _ = decoder_lstm[i](decoder_hidden, initial_state=encoder_states)
    
    # add linear layer on top of LSTM
    decoder_dense = Dense(n_features, activation=output_activation, name='dense_output')
    decoder_outputs = decoder_dense(decoder_hidden)
    
    # define seq2seq model
    model = Model([encoder_inputs, decoder_inputs], decoder_outputs)
    optimizer = Adam(lr=learning_rate)
    model.compile(optimizer=optimizer, loss=loss)
    
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
    
    return model, encoder_model, decoder_model