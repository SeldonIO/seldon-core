#from sklearn.pipeline import Pipeline
#from sklearn.base import BaseEstimator, ClassifierMixin
import numpy as np
import math
import datetime
#from seldon.pipeline import PipelineSaver
import os
import tensorflow as tf
from keras import backend
from keras.models import Model,load_model
from keras.layers import Dense,Input
from keras.layers import Dropout
from keras.layers import Flatten
from keras.constraints import maxnorm
from keras.layers.convolutional import Convolution2D
from keras.layers.convolutional import MaxPooling2D

from keras.callbacks import TensorBoard

class MnistFfnn(object):

    def __init__(self,
                 input_shape=(784,),
                 nb_labels=10,
                 optimizer='Adam',
                 run_dir='tensorboardlogs_test'):
        
        self.model_name='MnistFfnn'
        self.run_dir=run_dir
        self.input_shape=input_shape
        self.nb_labels=nb_labels
        self.optimizer=optimizer
        self.build_graph()

    def build_graph(self):
                            
        inp = Input(shape=self.input_shape,name='input_part')

        #keras layers
        with tf.name_scope('dense_1') as scope:
            h1 = Dense(256,
                         activation='relu',
                         W_constraint=maxnorm(3))(inp)
            drop1 = Dropout(0.2)(h1)

        with tf.name_scope('dense_2') as scope:
            h2 = Dense(128,
                       activation='relu',
                       W_constraint=maxnorm(3))(drop1)
            drop2 = Dropout(0.5)(h2)
            
            out = Dense(self.nb_labels,
                        activation='softmax')(drop2)

        self.model = Model(inp,out)
        
        if self.optimizer ==  'rmsprop':
            self.model.compile(loss='categorical_crossentropy',
                               optimizer='rmsprop',
                               metrics=['accuracy'])
        elif self.optimizer == 'Adam':
            self.model.compile(loss='categorical_crossentropy',
                               optimizer='Adam',
                               metrics=['accuracy'])
            
        print 'graph builded'

    def fit(self,X,y=None,
            X_test=None,y_test=None,
            batch_size=128,
            nb_epochs=2,
            shuffle=True):
        
        now = datetime.datetime.now()
        tensorboard_logname = self.run_dir+'/{}_{}'.format(self.model_name,
                                                           now.strftime('%Y.%m.%d_%H.%M'))      
        tensorboard = TensorBoard(log_dir=tensorboard_logname)
        
        self.model.fit(X,y,
                       validation_data=(X_test,y_test),
                       callbacks=[tensorboard],
                       batch_size=batch_size, 
                       nb_epoch=nb_epochs,
                       shuffle = shuffle)
        return self
    
    def predict_proba(self,X):

        return self.model.predict_proba(X)
    
    def predict(self, X):
        probas = self.model.predict_proba(X)
        return([[p>0.5 for p in p1] for p1 in probas])
        
    def score(self, X, y=None):
        pass

    def get_class_id_map(self):
        return ["proba"]

class MnistConv(object):

    def __init__(self,
                 input_shape=(28,28,1),
                 nb_labels=10,
                 optimizer='Adam',
                 run_dir='tensorboardlogs_test',
                 saved_model_file='MnistClassifier.h5'):
        
        self.model_name='MnistConv'
        self.run_dir=run_dir
        self.input_shape=input_shape
        self.nb_labels=nb_labels
        self.optimizer=optimizer
        self.saved_model_file=saved_model_file
        self.build_graph()

    def build_graph(self):
                            
        inp = Input(shape=self.input_shape,name='input_part')
        
        #keras layers
        with tf.name_scope('conv') as scope:
            conv = Convolution2D(32, 3, 3,
                                 input_shape=(32, 32, 3),
                                 border_mode='same',
                                 activation='relu',
                                 W_constraint=maxnorm(3))(inp)
            drop_conv = Dropout(0.2)(conv)
            max_pool = MaxPooling2D(pool_size=(2, 2))(drop_conv)

        with tf.name_scope('dense') as scope:
            flat = Flatten()(max_pool)                
            dense = Dense(128,
                          activation='relu',
                          W_constraint=maxnorm(3))(flat)
            drop_dense = Dropout(0.5)(dense)
            
            out = Dense(self.nb_labels,
                        activation='softmax')(drop_dense)

        self.model = Model(inp,out)
        
        if self.optimizer ==  'rmsprop':
            self.model.compile(loss='categorical_crossentropy',
                               optimizer='rmsprop',
                               metrics=['accuracy'])
        elif self.optimizer == 'Adam':
            self.model.compile(loss='categorical_crossentropy',
                               optimizer='Adam',
                               metrics=['accuracy'])
            
        print 'graph builded'

    def fit(self,X,y=None,
            X_test=None,y_test=None,
            batch_size=128,
            nb_epochs=2,
            shuffle=True):
        
        now = datetime.datetime.now()
        tensorboard_logname = self.run_dir+'/{}_{}'.format(self.model_name,
                                                           now.strftime('%Y.%m.%d_%H.%M'))      
        tensorboard = TensorBoard(log_dir=tensorboard_logname)
        
        self.model.fit(X,y,
                       validation_data=(X_test,y_test),
                       callbacks=[tensorboard],
                       batch_size=batch_size, 
                       nb_epoch=nb_epochs,
                       shuffle = shuffle)
        #if not os.path.exists('saved_model'):
        #    os.makedirs('saved_model')
        self.model.save(self.saved_model_file)
        return self
    
    def predict_proba(self,X):
        return self.model.predict_proba(X)
    
    def predict(self, X):
        probas = self.model.predict_proba(X)
        return([[p>0.5 for p in p1] for p1 in probas])
        
    def score(self, X, y=None):
        pass

    def get_class_id_map(self):
        return ["proba"]

from tensorflow.examples.tutorials.mnist import input_data
mnist = input_data.read_data_sets('data/MNIST_data', one_hot=True)
X_train = mnist.train.images
y_train = mnist.train.labels
X_test = mnist.test.images
y_test = mnist.test.labels

X_train = X_train.reshape((len(X_train),28,28,1))
X_test = X_test.reshape((len(X_test),28,28,1))

def main():
    mc = MnistConv((28,28,1),10)
    mc.fit(X_train,y=y_train,
           X_test=X_test,y_test=y_test)


if __name__ == "__main__":
    main()
