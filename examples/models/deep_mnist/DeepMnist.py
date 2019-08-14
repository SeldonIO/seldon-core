import tensorflow as tf
import numpy as np
import os

class DeepMnist(object):
    def __init__(self):
        self.loaded = False
        self.class_names = ["class:{}".format(str(i)) for i in range(10)]
        
    def load(self):
        print("Loading model",os.getpid())
        self.sess = tf.Session()
        saver = tf.train.import_meta_graph("model/deep_mnist_model.meta")
        saver.restore(self.sess,tf.train.latest_checkpoint("./model/"))
        graph = tf.get_default_graph()
        self.x = graph.get_tensor_by_name("x:0")
        self.y = graph.get_tensor_by_name("y:0")
        self.loaded = True
        print("Loaded model")
        
    def predict(self,X,feature_names):
        if not self.loaded:
            self.load()
        predictions = self.sess.run(self.y,feed_dict={self.x:X})
        return predictions.astype(np.float64)

    
