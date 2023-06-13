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
2023-03-09 17:05:36.275254: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2023-03-09 17:05:36.275268: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.

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
2023-03-09 17:05:46.653642: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcuda.so.1'; dlerror: libcuda.so.1: cannot open shared object file: No such file or directory; LD_LIBRARY_PATH: /home/clive/miniconda3/envs/scv2/lib/python3.9/site-packages/cv2/../../lib64:
2023-03-09 17:05:46.653676: W tensorflow/stream_executor/cuda/cuda_driver.cc:269] failed call to cuInit: UNKNOWN ERROR (303)
2023-03-09 17:05:46.653706: I tensorflow/stream_executor/cuda/cuda_diagnostics.cc:156] kernel driver does not appear to be running on this host (clive-ThinkPad-P1-Gen-5): /proc/driver/nvidia/version does not exist
2023-03-09 17:05:46.654247: I tensorflow/core/platform/cpu_feature_guard.cc:151] This TensorFlow binary is optimized with oneAPI Deep Neural Network Library (oneDNN) to use the following CPU instructions in performance-critical operations:  AVX2 FMA
To enable them in other operations, rebuild TensorFlow with the appropriate compiler flags.
2023-03-09 17:05:46.954048: W tensorflow/core/framework/cpu_allocator_impl.cc:82] Allocation of 614400000 exceeds 10% of free system memory.
2023-03-09 17:05:47.164088: W tensorflow/core/framework/cpu_allocator_impl.cc:82] Allocation of 614400000 exceeds 10% of free system memory.

```

```
782/782 [=] - 123s 157ms/step - loss_ma: 8047.0571

```

```
2023-03-09 17:07:50.493324: W tensorflow/core/framework/cpu_allocator_impl.cc:82] Allocation of 614400000 exceeds 10% of free system memory.

```

```
782/782 [=] - 121s 154ms/step - loss_ma: -2349.7290

```

```
2023-03-09 17:09:51.538684: W tensorflow/core/framework/cpu_allocator_impl.cc:82] Allocation of 614400000 exceeds 10% of free system memory.

```

```
782/782 [=] - 124s 158ms/step - loss_ma: -3547.2254

```

```
2023-03-09 17:11:55.269597: W tensorflow/core/framework/cpu_allocator_impl.cc:82] Allocation of 614400000 exceeds 10% of free system memory.

```

```
64/782 [.] - ETA: 1:50 - loss_ma: -3935.1125

```

```
---------------------------------------------------------------------------

```

```
KeyboardInterrupt                         Traceback (most recent call last)

```

```
Input In [3], in <cell line: 29>()
     22 od = OutlierVAE(threshold=.015,  # threshold for outlier score
     23                 score_type='mse',  # use MSE of reconstruction error for outlier detection
     24                 encoder_net=encoder_net,  # can also pass VAE model instead
     25                 decoder_net=decoder_net,  # of separate encoder and decoder
     26                 latent_dim=latent_dim,
     27                 samples=2)
     28 # train
---> 29 od.fit(X_train,
     30         loss_fn=elbo,
     31         cov_elbo=dict(sim=.05),
     32         epochs=50,
     33         verbose=True)
     35 # save the trained outlier detector
     36 save_detector(od, "./outlier-detector")

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/alibi_detect/od/vae.py:134, in OutlierVAE.fit(self, X, loss_fn, optimizer, cov_elbo, epochs, batch_size, verbose, log_metric, callbacks)
    131     kwargs['loss_fn_kwargs'] = {cov_elbo_type: tf.dtypes.cast(cov, tf.float32)}
    133 # train
--> 134 trainer(*args, **kwargs)

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/alibi_detect/models/tensorflow/trainer.py:93, in trainer(model, loss_fn, x_train, y_train, dataset, optimizer, loss_fn_kwargs, preprocess_fn, epochs, reg_loss_fn, batch_size, buffer_size, verbose, log_metric, callbacks)
     90         loss += sum(model.losses)
     91     loss += reg_loss_fn(model)  # alternative way they might be specified
---> 93 grads = tape.gradient(loss, model.trainable_weights)
     94 optimizer.apply_gradients(zip(grads, model.trainable_weights))
     95 if verbose:

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/tensorflow/python/eager/backprop.py:1081, in GradientTape.gradient(self, target, sources, output_gradients, unconnected_gradients)
   1077 if output_gradients is not None:
   1078   output_gradients = [None if x is None else ops.convert_to_tensor(x)
   1079                       for x in nest.flatten(output_gradients)]
-> 1081 flat_grad = imperative_grad.imperative_grad(
   1082     self._tape,
   1083     flat_targets,
   1084     flat_sources,
   1085     output_gradients=output_gradients,
   1086     sources_raw=flat_sources_raw,
   1087     unconnected_gradients=unconnected_gradients)
   1089 if not self._persistent:
   1090   # Keep track of watched variables before setting tape to None
   1091   self._watched_variables = self._tape.watched_variables()

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/tensorflow/python/eager/imperative_grad.py:67, in imperative_grad(tape, target, sources, output_gradients, sources_raw, unconnected_gradients)
     63 except ValueError:
     64   raise ValueError(
     65       "Unknown value for unconnected_gradients: %r" % unconnected_gradients)
---> 67 return pywrap_tfe.TFE_Py_TapeGradient(
     68     tape._tape,  # pylint: disable=protected-access
     69     target,
     70     sources,
     71     output_gradients,
     72     sources_raw,
     73     compat.as_str(unconnected_gradients.value))

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/tensorflow/python/eager/backprop.py:156, in _gradient_function(op_name, attr_tuple, num_inputs, inputs, outputs, out_grads, skip_input_indices, forward_pass_name_scope)
    154     gradient_name_scope += forward_pass_name_scope + "/"
    155   with ops.name_scope(gradient_name_scope):
--> 156     return grad_fn(mock_op, *out_grads)
    157 else:
    158   return grad_fn(mock_op, *out_grads)

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/tensorflow/python/ops/math_grad.py:1741, in _MatMulGrad(op, grad)
   1739 if not t_a and not t_b:
   1740   grad_a = gen_math_ops.mat_mul(grad, b, transpose_b=True)
-> 1741   grad_b = gen_math_ops.mat_mul(a, grad, transpose_a=True)
   1742 elif not t_a and t_b:
   1743   grad_a = gen_math_ops.mat_mul(grad, b)

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/tensorflow/python/ops/gen_math_ops.py:6013, in mat_mul(a, b, transpose_a, transpose_b, name)
   6011 if tld.is_eager:
   6012   try:
-> 6013     _result = pywrap_tfe.TFE_Py_FastPathExecute(
   6014       _ctx, "MatMul", name, a, b, "transpose_a", transpose_a, "transpose_b",
   6015       transpose_b)
   6016     return _result
   6017   except _core._NotOkStatusException as e:

```

```
KeyboardInterrupt:

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

