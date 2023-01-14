## Train Outlier Detector

Based on [Alibi Detect Example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/od_vae_cifar10.html).

```python
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

```

```
2022-10-02 18:02:43.572083: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2022-10-02 18:02:43.572120: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.

```

```python
train, test = tf.keras.datasets.cifar10.load_data()
X_train, y_train = train
X_test, y_test = test

X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
print(X_train.shape, y_train.shape, X_test.shape, y_test.shape)
```

```
(50000, 32, 32, 3) (50000, 1) (10000, 32, 32, 3) (10000, 1)

```

```python
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
```

```
782/782 [=] - 408s 521ms/step - loss_ma: 12246.5974
782/782 [=] - 441s 563ms/step - loss_ma: -336.8379
782/782 [=] - 514s 656ms/step - loss_ma: -2060.8556
782/782 [=] - 385s 491ms/step - loss_ma: -2942.0786
782/782 [=] - 334s 426ms/step - loss_ma: -3482.4651
782/782 [=] - 331s 422ms/step - loss_ma: -3826.7684
782/782 [=] - 328s 419ms/step - loss_ma: -4113.3548
782/782 [=] - 324s 414ms/step - loss_ma: -4344.3997
782/782 [=] - 356s 455ms/step - loss_ma: -4517.2223
782/782 [=] - 340s 434ms/step - loss_ma: -4636.2948
782/782 [=] - 338s 432ms/step - loss_ma: -4749.2001
782/782 [=] - 332s 423ms/step - loss_ma: -4853.0608
782/782 [=] - 333s 425ms/step - loss_ma: -4944.8810
782/782 [=] - 326s 416ms/step - loss_ma: -5008.7755
782/782 [=] - 326s 416ms/step - loss_ma: -5083.0938
782/782 [=] - 315s 402ms/step - loss_ma: -5140.1222
782/782 [=] - 311s 397ms/step - loss_ma: -5201.4831
782/782 [=] - 316s 403ms/step - loss_ma: -5250.6943
782/782 [=] - 318s 406ms/step - loss_ma: -5283.1387
782/782 [=] - 316s 404ms/step - loss_ma: -5321.9071
782/782 [=] - 320s 409ms/step - loss_ma: -5368.1527
782/782 [=] - 321s 409ms/step - loss_ma: -5388.1669
782/782 [=] - 317s 405ms/step - loss_ma: -5425.3241
782/782 [=] - 319s 408ms/step - loss_ma: -5449.0800
782/782 [=] - 320s 409ms/step - loss_ma: -5481.8549
782/782 [=] - 319s 407ms/step - loss_ma: -5511.5501
782/782 [=] - 318s 406ms/step - loss_ma: -5527.1451
782/782 [=] - 319s 408ms/step - loss_ma: -5554.7566
782/782 [=] - 373s 476ms/step - loss_ma: -5573.3425
782/782 [=] - 376s 480ms/step - loss_ma: -5599.4756
782/782 [=] - 375s 479ms/step - loss_ma: -5612.1058
782/782 [=] - 375s 478ms/step - loss_ma: -5622.7508
782/782 [=] - 373s 476ms/step - loss_ma: -5642.8896
782/782 [=] - 374s 477ms/step - loss_ma: -5661.9375
782/782 [=] - 373s 476ms/step - loss_ma: -5674.5788
782/782 [=] - 374s 477ms/step - loss_ma: -5683.5199
782/782 [=] - 373s 476ms/step - loss_ma: -5697.1596
782/782 [=] - 375s 479ms/step - loss_ma: -5711.3716
782/782 [=] - 376s 480ms/step - loss_ma: -5721.1513
782/782 [=] - 377s 481ms/step - loss_ma: -5733.8188
782/782 [=] - 376s 480ms/step - loss_ma: -5741.6455
782/782 [=] - 376s 480ms/step - loss_ma: -5743.0460
782/782 [=] - 376s 480ms/step - loss_ma: -5756.9715
782/782 [=] - 376s 480ms/step - loss_ma: -5761.5570
782/782 [=] - 376s 480ms/step - loss_ma: -5776.9497
782/782 [=] - 376s 479ms/step - loss_ma: -5783.7933
782/782 [=] - 374s 478ms/step - loss_ma: -5791.9981
782/782 [=] - 377s 481ms/step - loss_ma: -5801.7900
782/782 [=] - 377s 481ms/step - loss_ma: -5803.0183
782/782 [=] - 376s 480ms/step - loss_ma: -5816.9740

```

Create a MLServer model settings file: `model-settings.json`:

```json
{
  "name": "cifar10-outlier-detect",
  "implementation": "mlserver_alibi_detect.AlibiDetectRuntime",
  "parameters": {
    "uri": "./",
    "version": "v0.1.0"
  }
}

```

Save to local or remote storage the directory. Here we saved to Google Storage:

```bash
gsutil ls -R gs://seldon-models/mlserver/alibi-detect/cifar10-outlier

gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/:
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/OutlierVAE.dill
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/meta.dill
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model-settings.json
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/:
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/checkpoint
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/decoder_net.h5
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/encoder_net.h5
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/vae.ckpt.data-00000-of-00001
gs://seldon-models/mlserver/alibi-detect/cifar10-outlier/model/vae.ckpt.index

```


## Train Drift Detector

Based on [Alibi Detect Example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_ks_cifar10.html)

```python
import matplotlib.pyplot as plt
import numpy as np
import os
import tensorflow as tf

from alibi_detect.cd import KSDrift
from alibi_detect.models.tensorflow import scale_by_instance
from alibi_detect.utils.fetching import fetch_tf_model, fetch_detector
from alibi_detect.saving import save_detector, load_detector
from alibi_detect.datasets import fetch_cifar10c, corruption_types_cifar10c
```

```python
(X_train, y_train), (X_test, y_test) = tf.keras.datasets.cifar10.load_data()
X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
y_train = y_train.astype('int64').reshape(-1,)
y_test = y_test.astype('int64').reshape(-1,)
```

```python
np.random.seed(0)
n_test = X_test.shape[0]
idx = np.random.choice(n_test, size=n_test // 2, replace=False)
idx_h0 = np.delete(np.arange(n_test), idx, axis=0)
X_ref,y_ref = X_test[idx], y_test[idx]
X_h0, y_h0 = X_test[idx_h0], y_test[idx_h0]
print(X_ref.shape, X_h0.shape)
```

```
(5000, 32, 32, 3) (5000, 32, 32, 3)

```

```python
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
```

```
Directory drift-detector/preprocess_fn/model does not exist and is now created.

```

Create a MLServer model settings file: `model-settings.json`:

```json
{
  "name": "cifar10-drift",
  "implementation": "mlserver_alibi_detect.AlibiDetectRuntime",
  "parameters": {
    "uri": "./",
    "version": "v0.1.0"
  }
}

```

Save to local or remote storage the directory. Here we saved to Google Storage:

```bash
gsutil ls -R gs://seldon-models/mlserver/alibi-detect/cifar10-drift

gs://seldon-models/mlserver/alibi-detect/cifar10-drift/:
gs://seldon-models/mlserver/alibi-detect/cifar10-drift/
gs://seldon-models/mlserver/alibi-detect/cifar10-drift/KSDrift.dill
gs://seldon-models/mlserver/alibi-detect/cifar10-drift/meta.dill
gs://seldon-models/mlserver/alibi-detect/cifar10-drift/model-settings.json
gs://seldon-models/mlserver/alibi-detect/cifar10-drift/model/:
gs://seldon-models/mlserver/alibi-detect/cifar10-drift/model/encoder.h5

```

