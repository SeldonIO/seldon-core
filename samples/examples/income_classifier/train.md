## Train a tabular income classification model with monitoring and explanations

These steps are extracted from various Seldon Alibi notebooks

 * [Income data preparation and classifier](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_clf_adult.html)
 * [Drift detector](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_chi2ks_adult.html)
 * [Outlier detector](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/od_vae_adult.html)
 * [Explainer](https://docs.seldon.io/projects/alibi/en/stable/examples/anchor_tabular_adult.html)

```python
import os
import numpy as np
import pandas as pd
from typing import List, Tuple, Dict, Callable

from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.svm import LinearSVC

from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.compose import ColumnTransformer
from sklearn.pipeline import Pipeline
from sklearn.impute import SimpleImputer
from sklearn.metrics import accuracy_score
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from alibi.explainers import AnchorTabular

from alibi.datasets import fetch_adult
from alibi_detect.cd import ChiSquareDrift, TabularDrift

import tensorflow as tf
tf.keras.backend.clear_session()
from tensorflow.keras.layers import Dense, InputLayer

from alibi_detect.od import OutlierVAE
from alibi_detect.utils.perturbation import inject_outlier_tabular
```

```
2022-10-14 13:31:10.879606: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2022-10-14 13:31:10.879644: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.
2022-10-14 13:31:12.753639: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcuda.so.1'; dlerror: libnvidia-fatbinaryloader.so.396.37: cannot open shared object file: No such file or directory
2022-10-14 13:31:12.753696: W tensorflow/stream_executor/cuda/cuda_driver.cc:269] failed call to cuInit: UNKNOWN ERROR (303)
2022-10-14 13:31:12.753741: I tensorflow/stream_executor/cuda/cuda_diagnostics.cc:156] kernel driver does not appear to be running on this host (clive-T470p): /proc/driver/nvidia/version does not exist

```

## Fetch and prepare data

```python
# fetch adult dataset
adult = fetch_adult()

# separate columns in numerical and categorical.
categorical_names = [adult.feature_names[i] for i in adult.category_map.keys()]
categorical_ids = list(adult.category_map.keys())

numerical_names = [name for i, name in enumerate(adult.feature_names) if i not in adult.category_map.keys()]
numerical_ids = [i for i in range(len(adult.feature_names)) if i not in adult.category_map.keys()]

X = adult.data
Y = adult.target

feature_names = adult.feature_names
category_map = adult.category_map

labels = ['No!', 'Yes!']

def print_preds(preds: dict, preds_name: str) -> None:
    print(preds_name)
    print('Drift? {}'.format(labels[preds['data']['is_drift']]))
    print(f'p-value: {preds["data"]["p_val"]:.3f}')
    print('')
```

```python
education_col = adult.feature_names.index('Education')
education = adult.category_map[education_col]
print(education)
# define low education
low_education = [
    education.index('Dropout'),
    education.index('High School grad'),
    education.index('Bachelors')

]
# define high education
high_education = [
    education.index('Bachelors'),
    education.index('Masters'),
    education.index('Doctorate')
]
print("Low education:", [education[i] for i in low_education])
print("High education:", [education[i] for i in high_education])
# select instances for low and high education
low_education_mask = pd.Series(X[:, education_col]).isin(low_education).to_numpy()
high_education_mask = pd.Series(X[:, education_col]).isin(high_education).to_numpy()
X_low, X_high, Y_low, Y_high = X[low_education_mask], X[high_education_mask], Y[low_education_mask], Y[high_education_mask]
size = 1000
np.random.seed(0)

# define reference and H0 dataset
#idx_hgh = np.random.choice(np.arange(X_high.shape[0]), size=2*size, replace=False)
x_ref, x_h0, y_ref, y_h0 = train_test_split(X_high, Y_high, test_size=0.4, random_state=5, shuffle=True)

# define reference and H1 dataset
#idx_low = np.random.choice(np.arange(X_low.shape[0]), size=size, replace=False)
x_h1 = X_low
y_h1 = Y_low
```

```
['Associates', 'Bachelors', 'Doctorate', 'Dropout', 'High School grad', 'Masters', 'Prof-School']
Low education: ['Dropout', 'High School grad', 'Bachelors']
High education: ['Bachelors', 'Masters', 'Doctorate']

```

## Drift Detector

```python
categories_per_feature = {f: None for f in list(category_map.keys())}
cd = TabularDrift(x_ref, p_val=.05, categories_per_feature=categories_per_feature)
```

```python
preds = cd.predict(x_h0)
print('Drift? {}'.format(labels[preds['data']['is_drift']]))
```

```
Drift? No!

```

```python
preds = cd.predict(x_h1)
print('Drift? {}'.format(labels[preds['data']['is_drift']]))
```

```
Drift? Yes!

```

```python
from alibi_detect.utils.saving import save_detector
save_detector(cd, "./drift-detector")
```

## Model

```python
ordinal_features = [x for x in range(len(feature_names)) if x not in list(category_map.keys())]
ordinal_transformer = Pipeline(steps=[('imputer', SimpleImputer(strategy='median')),
                                      ('scaler', StandardScaler())])
categorical_features = list(category_map.keys())
categorical_transformer = Pipeline(steps=[('imputer', SimpleImputer(strategy='median')),
                                          ('onehot', OneHotEncoder(handle_unknown='ignore'))])
preprocessor = ColumnTransformer(transformers=[('num', ordinal_transformer, ordinal_features),
                                               ('cat', categorical_transformer, categorical_features)],
                                sparse_threshold=0)
np.random.seed(0)
clf = RandomForestClassifier(n_estimators=50)

train_pipeline = Pipeline(steps=[('preprocessor',preprocessor),('classifier',clf)])
train_pipeline.fit(x_ref, y_ref)
```

<style>#sk-container-id-1 {color: black;background-color: white;}#sk-container-id-1 pre{padding: 0;}#sk-container-id-1 div.sk-toggleable {background-color: white;}#sk-container-id-1 label.sk-toggleable__label {cursor: pointer;display: block;width: 100%;margin-bottom: 0;padding: 0.3em;box-sizing: border-box;text-align: center;}#sk-container-id-1 label.sk-toggleable__label-arrow:before {content: "▸";float: left;margin-right: 0.25em;color: #696969;}#sk-container-id-1 label.sk-toggleable__label-arrow:hover:before {color: black;}#sk-container-id-1 div.sk-estimator:hover label.sk-toggleable__label-arrow:before {color: black;}#sk-container-id-1 div.sk-toggleable__content {max-height: 0;max-width: 0;overflow: hidden;text-align: left;background-color: #f0f8ff;}#sk-container-id-1 div.sk-toggleable__content pre {margin: 0.2em;color: black;border-radius: 0.25em;background-color: #f0f8ff;}#sk-container-id-1 input.sk-toggleable__control:checked~div.sk-toggleable__content {max-height: 200px;max-width: 100%;overflow: auto;}#sk-container-id-1 input.sk-toggleable__control:checked~label.sk-toggleable__label-arrow:before {content: "▾";}#sk-container-id-1 div.sk-estimator input.sk-toggleable__control:checked~label.sk-toggleable__label {background-color: #d4ebff;}#sk-container-id-1 div.sk-label input.sk-toggleable__control:checked~label.sk-toggleable__label {background-color: #d4ebff;}#sk-container-id-1 input.sk-hidden--visually {border: 0;clip: rect(1px 1px 1px 1px);clip: rect(1px, 1px, 1px, 1px);height: 1px;margin: -1px;overflow: hidden;padding: 0;position: absolute;width: 1px;}#sk-container-id-1 div.sk-estimator {font-family: monospace;background-color: #f0f8ff;border: 1px dotted black;border-radius: 0.25em;box-sizing: border-box;margin-bottom: 0.5em;}#sk-container-id-1 div.sk-estimator:hover {background-color: #d4ebff;}#sk-container-id-1 div.sk-parallel-item::after {content: "";width: 100%;border-bottom: 1px solid gray;flex-grow: 1;}#sk-container-id-1 div.sk-label:hover label.sk-toggleable__label {background-color: #d4ebff;}#sk-container-id-1 div.sk-serial::before {content: "";position: absolute;border-left: 1px solid gray;box-sizing: border-box;top: 0;bottom: 0;left: 50%;z-index: 0;}#sk-container-id-1 div.sk-serial {display: flex;flex-direction: column;align-items: center;background-color: white;padding-right: 0.2em;padding-left: 0.2em;position: relative;}#sk-container-id-1 div.sk-item {position: relative;z-index: 1;}#sk-container-id-1 div.sk-parallel {display: flex;align-items: stretch;justify-content: center;background-color: white;position: relative;}#sk-container-id-1 div.sk-item::before, #sk-container-id-1 div.sk-parallel-item::before {content: "";position: absolute;border-left: 1px solid gray;box-sizing: border-box;top: 0;bottom: 0;left: 50%;z-index: -1;}#sk-container-id-1 div.sk-parallel-item {display: flex;flex-direction: column;z-index: 1;position: relative;background-color: white;}#sk-container-id-1 div.sk-parallel-item:first-child::after {align-self: flex-end;width: 50%;}#sk-container-id-1 div.sk-parallel-item:last-child::after {align-self: flex-start;width: 50%;}#sk-container-id-1 div.sk-parallel-item:only-child::after {width: 0;}#sk-container-id-1 div.sk-dashed-wrapped {border: 1px dashed gray;margin: 0 0.4em 0.5em 0.4em;box-sizing: border-box;padding-bottom: 0.4em;background-color: white;}#sk-container-id-1 div.sk-label label {font-family: monospace;font-weight: bold;display: inline-block;line-height: 1.2em;}#sk-container-id-1 div.sk-label-container {text-align: center;}#sk-container-id-1 div.sk-container {/* jupyter's `normalize.less` sets `[hidden] { display: none; }` but bootstrap.min.css set `[hidden] { display: none !important; }` so we also need the `!important` here to be able to override the default hidden behavior on the sphinx rendered scikit-learn.org. See: https://github.com/scikit-learn/scikit-learn/issues/21755 */display: inline-block !important;position: relative;}#sk-container-id-1 div.sk-text-repr-fallback {display: none;}</style><div id="sk-container-id-1" class="sk-top-container"><div class="sk-text-repr-fallback"><pre>Pipeline(steps=[(&#x27;preprocessor&#x27;,
```
ColumnTransformer(sparse_threshold=0,
                               transformers=[(&#x27;num&#x27;,
                                              Pipeline(steps=[(&#x27;imputer&#x27;,
                                                               SimpleImputer(strategy=&#x27;median&#x27;)),
                                                              (&#x27;scaler&#x27;,
                                                               StandardScaler())]),
                                              [0, 8, 9, 10]),
                                             (&#x27;cat&#x27;,
                                              Pipeline(steps=[(&#x27;imputer&#x27;,
                                                               SimpleImputer(strategy=&#x27;median&#x27;)),
                                                              (&#x27;onehot&#x27;,
                                                               OneHotEncoder(handle_unknown=&#x27;ignore&#x27;))]),
                                              [1, 2, 3, 4, 5, 6, 7, 11])])),
            (&#x27;classifier&#x27;, RandomForestClassifier(n_estimators=50))])</pre><b>In a Jupyter environment, please rerun this cell to show the HTML representation or trust the notebook. <br />On GitHub, the HTML representation is unable to render, please try loading this page with nbviewer.org.</b></div><div class="sk-container" hidden><div class="sk-item sk-dashed-wrapped"><div class="sk-label-container"><div class="sk-label sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-1" type="checkbox" ><label for="sk-estimator-id-1" class="sk-toggleable__label sk-toggleable__label-arrow">Pipeline</label><div class="sk-toggleable__content"><pre>Pipeline(steps=[(&#x27;preprocessor&#x27;,
             ColumnTransformer(sparse_threshold=0,
                               transformers=[(&#x27;num&#x27;,
                                              Pipeline(steps=[(&#x27;imputer&#x27;,
                                                               SimpleImputer(strategy=&#x27;median&#x27;)),
                                                              (&#x27;scaler&#x27;,
                                                               StandardScaler())]),
                                              [0, 8, 9, 10]),
                                             (&#x27;cat&#x27;,
                                              Pipeline(steps=[(&#x27;imputer&#x27;,
                                                               SimpleImputer(strategy=&#x27;median&#x27;)),
                                                              (&#x27;onehot&#x27;,
                                                               OneHotEncoder(handle_unknown=&#x27;ignore&#x27;))]),
                                              [1, 2, 3, 4, 5, 6, 7, 11])])),
            (&#x27;classifier&#x27;, RandomForestClassifier(n_estimators=50))])</pre></div></div></div><div class="sk-serial"><div class="sk-item sk-dashed-wrapped"><div class="sk-label-container"><div class="sk-label sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-2" type="checkbox" ><label for="sk-estimator-id-2" class="sk-toggleable__label sk-toggleable__label-arrow">preprocessor: ColumnTransformer</label><div class="sk-toggleable__content"><pre>ColumnTransformer(sparse_threshold=0,
              transformers=[(&#x27;num&#x27;,
                             Pipeline(steps=[(&#x27;imputer&#x27;,
                                              SimpleImputer(strategy=&#x27;median&#x27;)),
                                             (&#x27;scaler&#x27;, StandardScaler())]),
                             [0, 8, 9, 10]),
                            (&#x27;cat&#x27;,
                             Pipeline(steps=[(&#x27;imputer&#x27;,
                                              SimpleImputer(strategy=&#x27;median&#x27;)),
                                             (&#x27;onehot&#x27;,
                                              OneHotEncoder(handle_unknown=&#x27;ignore&#x27;))]),
                             [1, 2, 3, 4, 5, 6, 7, 11])])</pre></div></div></div><div class="sk-parallel"><div class="sk-parallel-item"><div class="sk-item"><div class="sk-label-container"><div class="sk-label sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-3" type="checkbox" ><label for="sk-estimator-id-3" class="sk-toggleable__label sk-toggleable__label-arrow">num</label><div class="sk-toggleable__content"><pre>[0, 8, 9, 10]</pre></div></div></div><div class="sk-serial"><div class="sk-item"><div class="sk-serial"><div class="sk-item"><div class="sk-estimator sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-4" type="checkbox" ><label for="sk-estimator-id-4" class="sk-toggleable__label sk-toggleable__label-arrow">SimpleImputer</label><div class="sk-toggleable__content"><pre>SimpleImputer(strategy=&#x27;median&#x27;)</pre></div></div></div><div class="sk-item"><div class="sk-estimator sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-5" type="checkbox" ><label for="sk-estimator-id-5" class="sk-toggleable__label sk-toggleable__label-arrow">StandardScaler</label><div class="sk-toggleable__content"><pre>StandardScaler()</pre></div></div></div></div></div></div></div></div><div class="sk-parallel-item"><div class="sk-item"><div class="sk-label-container"><div class="sk-label sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-6" type="checkbox" ><label for="sk-estimator-id-6" class="sk-toggleable__label sk-toggleable__label-arrow">cat</label><div class="sk-toggleable__content"><pre>[1, 2, 3, 4, 5, 6, 7, 11]</pre></div></div></div><div class="sk-serial"><div class="sk-item"><div class="sk-serial"><div class="sk-item"><div class="sk-estimator sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-7" type="checkbox" ><label for="sk-estimator-id-7" class="sk-toggleable__label sk-toggleable__label-arrow">SimpleImputer</label><div class="sk-toggleable__content"><pre>SimpleImputer(strategy=&#x27;median&#x27;)</pre></div></div></div><div class="sk-item"><div class="sk-estimator sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-8" type="checkbox" ><label for="sk-estimator-id-8" class="sk-toggleable__label sk-toggleable__label-arrow">OneHotEncoder</label><div class="sk-toggleable__content"><pre>OneHotEncoder(handle_unknown=&#x27;ignore&#x27;)</pre></div></div></div></div></div></div></div></div></div></div><div class="sk-item"><div class="sk-estimator sk-toggleable"><input class="sk-toggleable__control sk-hidden--visually" id="sk-estimator-id-9" type="checkbox" ><label for="sk-estimator-id-9" class="sk-toggleable__label sk-toggleable__label-arrow">RandomForestClassifier</label><div class="sk-toggleable__content"><pre>RandomForestClassifier(n_estimators=50)</pre></div></div></div></div></div></div></div>

```

```python
predict_fn = lambda x: train_pipeline.predict(x)
print('Train accuracy: ', accuracy_score(y_ref, predict_fn(x_ref)))
print('Test accuracy: ', accuracy_score(y_h0, predict_fn(x_h0)))
```

```
Train accuracy:  0.983756119270138
Test accuracy:  0.7774441107774441

```

```python
from joblib import dump, load
os.makedirs("./classifier", exist_ok=True)
dump(train_pipeline, './classifier/model.joblib')
```

```
['./classifier/model.joblib']

```

## Outlier Detector

```python
os.makedirs("./preprocessor", exist_ok=True)
dump(preprocessor, './preprocessor/model.joblib')
```

```
['./preprocessor/model.joblib']

```

```python
X_train = preprocessor.transform(x_ref)
```

```python
n_features = X_train.shape[1]
latent_dim = 2

encoder_net = tf.keras.Sequential(
    [
        InputLayer(input_shape=(n_features,)),
        Dense(25, activation=tf.nn.relu),
         Dense(10, activation=tf.nn.relu),
        Dense(5, activation=tf.nn.relu)
    ])

decoder_net = tf.keras.Sequential(
    [
        InputLayer(input_shape=(latent_dim,)),
        Dense(5, activation=tf.nn.relu),
        Dense(10, activation=tf.nn.relu),
        Dense(25, activation=tf.nn.relu),
        Dense(n_features, activation=None)
    ])

# initialize outlier detector
od = OutlierVAE(threshold=None,  # threshold for outlier score
                score_type='mse',  # use MSE of reconstruction error for outlier detection
                encoder_net=encoder_net,  # can also pass VAE model instead
                decoder_net=decoder_net,  # of separate encoder and decoder
                latent_dim=latent_dim,
                samples=5)

# train
od.fit(X_train,
        loss_fn=tf.keras.losses.mse,
         epochs=5,
        verbose=True)
```

```
2022-10-14 13:32:16.370485: I tensorflow/core/platform/cpu_feature_guard.cc:151] This TensorFlow binary is optimized with oneAPI Deep Neural Network Library (oneDNN) to use the following CPU instructions in performance-critical operations:  AVX2 FMA
To enable them in other operations, rebuild TensorFlow with the appropriate compiler flags.
No threshold level set. Need to infer threshold using `infer_threshold`.

```

```
71/71 [=] - 1s 13ms/step - loss_ma: 0.2847
71/71 [=] - 1s 13ms/step - loss_ma: 0.1822
71/71 [=] - 1s 13ms/step - loss_ma: 0.1627
71/71 [=] - 1s 13ms/step - loss_ma: 0.1605
71/71 [=] - 1s 13ms/step - loss_ma: 0.1547

```

```python
cat_cols = list(category_map.keys())
num_cols = [col for col in range(x_ref.shape[1]) if col not in cat_cols]
print(cat_cols, num_cols)
```

```
[1, 2, 3, 4, 5, 6, 7, 11] [0, 8, 9, 10]

```

```python
perc_outlier = 10
data = inject_outlier_tabular(x_ref, num_cols, perc_outlier, n_std=8., min_std=6.)
X_threshold, y_threshold = data.data, data.target
X_threshold_, y_threshold_ = X_threshold.copy(), y_threshold.copy()  # store for comparison later
outlier_perc = 100 * y_threshold.sum() / len(y_threshold)
print('{:.2f}% outliers'.format(outlier_perc))
```

```
9.70% outliers

```

```python
perc_outlier = 100
data = inject_outlier_tabular(x_ref, num_cols, perc_outlier, n_std=8., min_std=6.)
X_outliers, y_outliers = data.data, data.target
```

```python
v = np.c_[preprocessor.transform(X_threshold)]
od.infer_threshold(v, threshold_perc=100-outlier_perc, outlier_perc=100)
print('New threshold: {}'.format(od.threshold))
```

```
New threshold: 0.7037604962160285

```

```python
save_detector(od, "./outlier-detector")
```

```
WARNING:tensorflow:Compiled the loaded model, but the compiled metrics have yet to be built. `model.compile_metrics` will be empty until you train or evaluate the model.
WARNING:tensorflow:Compiled the loaded model, but the compiled metrics have yet to be built. `model.compile_metrics` will be empty until you train or evaluate the model.

```

## Explainer

```python
predict_fn = lambda x: train_pipeline.predict(x)
explainer = AnchorTabular(predict_fn, feature_names, categorical_names=category_map, seed=1)
explainer.fit(x_ref, disc_perc=[25, 50, 75])
```

```
AnchorTabular(meta={
  'name': 'AnchorTabular',
  'type': ['blackbox'],
  'explanations': ['local'],
  'params': {'seed': 1, 'disc_perc': [25, 50, 75]},
  'version': '0.8.0'}
)

```

```python
idx = 0
class_names = adult.target_names
print('Prediction: ', class_names[explainer.predictor(x_h0[idx].reshape(1, -1))[0]])
```

```
Prediction:  <=50K

```

```python
explanation = explainer.explain(x_h0[idx], threshold=0.95)
print('Anchor: %s' % (' AND '.join(explanation.anchor)))
print('Precision: %.2f' % explanation.precision)
print('Coverage: %.2f' % explanation.coverage)
```

```
Could not find an anchor satisfying the 0.95 precision constraint. Now returning the best non-eligible result. The desired precision threshold might not be achieved due to the quantile-based discretisation of the numerical features. The resolution of the bins may be too large to find an anchor of required precision. Consider increasing the number of bins in `disc_perc`, but note that for some numerical distribution (e.g. skewed distribution) it may not help.

```

```yaml
Anchor: Age <= 31.00 AND Hours per week <= 40.00 AND Capital Gain <= 0.00 AND Education = Bachelors AND Capital Loss <= 0.00 AND Workclass = Private AND Country = United-States
Precision: 0.88
Coverage: 0.10

```

```python
from alibi.saving import save_explainer
save_explainer(explainer,"./explainer/data")
```

## Save Data

```python
os.makedirs("./infer-data", exist_ok=True)
with open('./infer-data/test.npy', 'wb') as f:
    np.save(f,x_ref)
    np.save(f,x_h1)
    np.save(f,y_ref)
    np.save(f,X_outliers)
```

```python

```
