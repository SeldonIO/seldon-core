# Copyright (c) 2024 Seldon Technologies Ltd.

# Use of this software is governed BY
# (1) the license included in the LICENSE file or
# (2) if the license included in the LICENSE file is the Business Source License 1.1,
# the Change License after the Change Date as each is defined in accordance with the LICENSE file.

#!/usr/bin/env python
# coding: utf-8

# ## Train Outlier Detector
# 
# Based on [Alibi Detect Example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/od_vae_cifar10.html).

# In[1]:


import os
import logging
import matplotlib.pyplot as plt
import numpy as np
import tensorflow as tf
tf.keras.backend.clear_session()
from tensorflow.keras.layers import Conv2D, Conv2DTranspose, Dense, Layer, Reshape, InputLayer
from tqdm import tqdm

from alibi_detect.models.tensorflow import elbo
from alibi_detect.od import OutlierVAE
from alibi_detect.utils.fetching import fetch_detector
from alibi_detect.utils.perturbation import apply_mask
from alibi_detect.saving import save_detector, load_detector
from alibi_detect.utils.visualize import plot_instance_score, plot_feature_outlier_image

logger = tf.get_logger()
logger.setLevel(logging.ERROR)


# In[2]:


train, test = tf.keras.datasets.cifar10.load_data()
X_train, y_train = train
X_test, y_test = test

X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
print(X_train.shape, y_train.shape, X_test.shape, y_test.shape)


# In[ ]:


latent_dim = 1024

encoder_net = tf.keras.Sequential(
    [
        InputLayer(input_shape=(32, 32, 3)),
        Conv2D(64, 4, strides=2, padding='same', activation=tf.nn.relu),
        Conv2D(128, 4, strides=2, padding='same', activation=tf.nn.relu),
        Conv2D(512, 4, strides=2, padding='same', activation=tf.nn.relu)
    ])

decoder_net = tf.keras.Sequential(
    [
        InputLayer(input_shape=(latent_dim,)),
        Dense(4*4*128),
        Reshape(target_shape=(4, 4, 128)),
        Conv2DTranspose(256, 4, strides=2, padding='same', activation=tf.nn.relu),
        Conv2DTranspose(64, 4, strides=2, padding='same', activation=tf.nn.relu),
        Conv2DTranspose(3, 4, strides=2, padding='same', activation='sigmoid')
    ])

# initialize outlier detector
od = OutlierVAE(threshold=.015,  # threshold for outlier score
                score_type='mse',  # use MSE of reconstruction error for outlier detection
                encoder_net=encoder_net,  # can also pass VAE model instead
                decoder_net=decoder_net,  # of separate encoder and decoder
                latent_dim=latent_dim,
                samples=2)
# train
od.fit(X_train,
        loss_fn=elbo,
        cov_elbo=dict(sim=.05),
        epochs=50,
        verbose=True)

# save the trained outlier detector
save_detector(od, "./outlier-detector")


# Create a MLServer model settings file: `model-settings.json`:
# 
# ```json
# {
#   "name": "cifar10-outlier-detect",
#   "implementation": "mlserver_alibi_detect.AlibiDetectRuntime",
#   "parameters": {
#     "uri": "./",
#     "version": "v0.1.0"
#   }
# }
# 
# ```
# 
# Save to local or remote storage the directory. Here we saved to Google Storage:
# 
# ```bash
# gsutil ls -R gs://seldon-models/mlserver/alibi-detect/cifar10-outlier 
# 
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/:
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/OutlierVAE.dill
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/meta.dill
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model-settings.json
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/:
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/checkpoint
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/decoder_net.h5
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/encoder_net.h5
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/vae.ckpt.data-00000-of-00001
# gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/vae.ckpt.index
# 
# ```
#  

# ## Train Drift Detector
# 
# Based on [Alibi Detect Example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_ks_cifar10.html)

# In[5]:


import matplotlib.pyplot as plt
import numpy as np
import os
import tensorflow as tf

from alibi_detect.cd import KSDrift
from alibi_detect.models.tensorflow import scale_by_instance
from alibi_detect.utils.fetching import fetch_tf_model, fetch_detector
from alibi_detect.saving import save_detector, load_detector
from alibi_detect.datasets import fetch_cifar10c, corruption_types_cifar10c


# In[6]:


(X_train, y_train), (X_test, y_test) = tf.keras.datasets.cifar10.load_data()
X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
y_train = y_train.astype('int64').reshape(-1,)
y_test = y_test.astype('int64').reshape(-1,)


# In[8]:


np.random.seed(0)
n_test = X_test.shape[0]
idx = np.random.choice(n_test, size=n_test // 2, replace=False)
idx_h0 = np.delete(np.arange(n_test), idx, axis=0)
X_ref,y_ref = X_test[idx], y_test[idx]
X_h0, y_h0 = X_test[idx_h0], y_test[idx_h0]
print(X_ref.shape, X_h0.shape)


# In[10]:


from functools import partial
from tensorflow.keras.layers import Conv2D, Dense, Flatten, InputLayer, Reshape
from alibi_detect.cd.tensorflow import preprocess_drift

tf.random.set_seed(0)

# define encoder
encoding_dim = 32
encoder_net = tf.keras.Sequential(
  [
      InputLayer(input_shape=(32, 32, 3)),
      Conv2D(64, 4, strides=2, padding='same', activation=tf.nn.relu),
      Conv2D(128, 4, strides=2, padding='same', activation=tf.nn.relu),
      Conv2D(512, 4, strides=2, padding='same', activation=tf.nn.relu),
      Flatten(),
      Dense(encoding_dim,)
  ]
)

# define preprocessing function
preprocess_fn = partial(preprocess_drift, model=encoder_net, batch_size=512)

# initialise drift detector
p_val = .05
cd = KSDrift(X_ref, p_val=p_val, preprocess_fn=preprocess_fn)

# we can also save/load an initialised detector
filepath = 'my_path'  # change to directory where detector is saved
save_detector(cd, "./drift-detector")


# Create a MLServer model settings file: `model-settings.json`:
# 
# ```json
# {
#   "name": "cifar10-drift",
#   "implementation": "mlserver_alibi_detect.AlibiDetectRuntime",
#   "parameters": {
#     "uri": "./",
#     "version": "v0.1.0"
#   }
# }
# ```
# 
# Save to local or remote storage the directory. Here we saved to Google Storage:
# 
# ```bash
# gsutil ls -R gs://seldon-models/mlserver/alibi-detect/cifar10-drift 
# 
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/:
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/KSDrift.dill
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/meta.dill
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/model-settings.json
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/model/:
# gs://seldon-models/mlserver/alibi-detect/cifar10-drift/model/encoder.h5
# 
# ```
#  
