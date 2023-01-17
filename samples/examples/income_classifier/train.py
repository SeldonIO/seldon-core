#!/usr/bin/env python
# coding: utf-8

# ## Train a tabular income classification model with monitoring and explanations
# 
# These steps are extracted from various Seldon Alibi notebooks
# 
#  * [Income data preparation and classifier](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_clf_adult.html)
#  * [Drift detector](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_chi2ks_adult.html)
#  * [Outlier detector](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/od_vae_adult.html)
#  * [Explainer](https://docs.seldon.io/projects/alibi/en/stable/examples/anchor_tabular_adult.html)

# In[1]:


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


# ## Fetch and prepare data

# In[4]:


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


# In[5]:


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


# ## Drift Detector

# In[6]:


categories_per_feature = {f: None for f in list(category_map.keys())}
cd = TabularDrift(x_ref, p_val=.05, categories_per_feature=categories_per_feature)


# In[7]:


preds = cd.predict(x_h0)
print('Drift? {}'.format(labels[preds['data']['is_drift']]))


# In[8]:


preds = cd.predict(x_h1)
print('Drift? {}'.format(labels[preds['data']['is_drift']]))


# In[26]:


from alibi_detect.utils.saving import save_detector
save_detector(cd, "./drift-detector")


# ## Model

# In[10]:


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


# In[11]:


predict_fn = lambda x: train_pipeline.predict(x)
print('Train accuracy: ', accuracy_score(y_ref, predict_fn(x_ref)))
print('Test accuracy: ', accuracy_score(y_h0, predict_fn(x_h0)))


# In[12]:


from joblib import dump, load
os.makedirs("./classifier", exist_ok=True)
dump(train_pipeline, './classifier/model.joblib') 


# ## Outlier Detector

# In[13]:


os.makedirs("./preprocessor", exist_ok=True)
dump(preprocessor, './preprocessor/model.joblib') 


# In[14]:


X_train = preprocessor.transform(x_ref)


# In[15]:


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


# In[16]:


cat_cols = list(category_map.keys())
num_cols = [col for col in range(x_ref.shape[1]) if col not in cat_cols]
print(cat_cols, num_cols)


# In[17]:


perc_outlier = 10
data = inject_outlier_tabular(x_ref, num_cols, perc_outlier, n_std=8., min_std=6.)
X_threshold, y_threshold = data.data, data.target
X_threshold_, y_threshold_ = X_threshold.copy(), y_threshold.copy()  # store for comparison later
outlier_perc = 100 * y_threshold.sum() / len(y_threshold)
print('{:.2f}% outliers'.format(outlier_perc))


# In[18]:


perc_outlier = 100
data = inject_outlier_tabular(x_ref, num_cols, perc_outlier, n_std=8., min_std=6.)
X_outliers, y_outliers = data.data, data.target


# In[19]:


v = np.c_[preprocessor.transform(X_threshold)]
od.infer_threshold(v, threshold_perc=100-outlier_perc, outlier_perc=100)
print('New threshold: {}'.format(od.threshold))


# In[20]:


save_detector(od, "./outlier-detector")


# ## Explainer

# In[21]:


predict_fn = lambda x: train_pipeline.predict(x)
explainer = AnchorTabular(predict_fn, feature_names, categorical_names=category_map, seed=1)
explainer.fit(x_ref, disc_perc=[25, 50, 75])


# In[22]:


idx = 0
class_names = adult.target_names
print('Prediction: ', class_names[explainer.predictor(x_h0[idx].reshape(1, -1))[0]])


# In[23]:


explanation = explainer.explain(x_h0[idx], threshold=0.95)
print('Anchor: %s' % (' AND '.join(explanation.anchor)))
print('Precision: %.2f' % explanation.precision)
print('Coverage: %.2f' % explanation.coverage)


# In[24]:


from alibi.saving import save_explainer
save_explainer(explainer,"./explainer/data")


# ## Save Data

# In[25]:


os.makedirs("./infer-data", exist_ok=True)
with open('./infer-data/test.npy', 'wb') as f:
    np.save(f,x_ref)
    np.save(f,x_h1)
    np.save(f,y_ref)
    np.save(f,X_outliers)


# In[ ]:




