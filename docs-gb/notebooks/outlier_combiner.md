# Seldon deployment of Alibi Outlier detector

Tne objective of this tutorial is to build a "loan approval" classifier equipped with the outliers detector from [alibi-detect](https://github.com/SeldonIO/alibi-detect) package.
The diagram of this tutorial is as follows:

In this tutorial we will follow the following steps:

1) Train and test model to predict loan approvals

2) Train and test outliers detector

3) Containerise and deploy your models

4) Test your your new seldon deployment


### Before you start
Make sure you install the following dependencies, as they are critical for this example to work:

* Helm v3.0.0+
* A Kubernetes cluster running v1.13 or above (minkube / docker-for-windows work well if enough RAM)
* kubectl v1.14+
* ksonnet v0.13.1+
* kfctl 0.5.1 - Please use this exact version as there are major changes every few months
* Python 3.6+
* Python DEV requirements (we'll install them below)

You can follow this [notebook](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) to setup your cluster.

Let's get started! 🚀🔥 


### Install Python dependencies

In [requirements-dev](https://github.com/SeldonIO/seldon-core/blob/master/examples/outliers/alibi-detect-combiner/requirements-dev.txt) file you will find set of python dependencies required to run this notebook.


```python
!cat requirements-dev.txt
```

    dill==0.3.1
    xai==0.0.5
    
    alibi==0.3.2
    alibi-detect==0.2.0
    seldon_core==1.0
    
    scipy==1.1.0
    numpy==1.15.4
    scikit-learn==0.20.1



```python
!pip install -r requirements-dev.txt
```

## Train and test loanclassifier

We start with training the loanclassifier model by using a prepared python script [train_classifier](https://github.com/SeldonIO/seldon-core/blob/master/examples/outliers/alibi-detect-combiner/train_classifier.py):


```python
!pygmentize train_classifier.py
```

    [34mimport[39;49;00m [04m[36malibi[39;49;00m
    [34mimport[39;49;00m [04m[36mdill[39;49;00m
    [34mimport[39;49;00m [04m[36mnumpy[39;49;00m [34mas[39;49;00m [04m[36mnp[39;49;00m
    
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mpreprocessing[39;49;00m [34mimport[39;49;00m StandardScaler, OneHotEncoder
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mimpute[39;49;00m [34mimport[39;49;00m SimpleImputer
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mpipeline[39;49;00m [34mimport[39;49;00m Pipeline
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mcompose[39;49;00m [34mimport[39;49;00m ColumnTransformer
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mensemble[39;49;00m [34mimport[39;49;00m RandomForestClassifier
    
    
    DATA_DIR = [33m"[39;49;00m[33mpipeline/loanclassifier[39;49;00m[33m"[39;49;00m
    
    
    [34mdef[39;49;00m [32mload_data[39;49;00m(train_size=[34m30000[39;49;00m, random_state=[34m0[39;49;00m):
        [33m"""Load example dataset and split between train and test datasets."""[39;49;00m
        [36mprint[39;49;00m([33m"[39;49;00m[33mLoading adult data from alibi.[39;49;00m[33m"[39;49;00m)
        np.random.seed(random_state)
    
        data = alibi.datasets.fetch_adult()
    
        [37m# mix input data[39;49;00m
        data_perm = np.random.permutation(np.c_[data.data, data.target])
        data.data = data_perm[:, :-[34m1[39;49;00m]
        data.target = data_perm[:, -[34m1[39;49;00m]
    
        [37m# perform train / test split[39;49;00m
        X_train, y_train = data.data[:train_size, :], data.target[:train_size]
        X_test, y_test = data.data[train_size:, :], data.target[train_size:]
    
        [34mreturn[39;49;00m data, X_train, y_train, X_test, y_test
    
    
    [34mdef[39;49;00m [32mtrain_preprocessor[39;49;00m(data):
        [33m"""Train preprocessor."""[39;49;00m
        [36mprint[39;49;00m([33m"[39;49;00m[33mTraining preprocessor.[39;49;00m[33m"[39;49;00m)
        [37m# TODO: ask if we need np.random.seed(...) here[39;49;00m
    
        ordinal_features = [
            n [34mfor[39;49;00m (n, _) [35min[39;49;00m [36menumerate[39;49;00m(data.feature_names)
            [34mif[39;49;00m n [35mnot[39;49;00m [35min[39;49;00m data.category_map
        ]
    
        categorical_features = [36mlist[39;49;00m(data.category_map.keys())
        ordinal_transformer = Pipeline(steps=[
                ([33m'[39;49;00m[33mimputer[39;49;00m[33m'[39;49;00m, SimpleImputer(strategy=[33m'[39;49;00m[33mmedian[39;49;00m[33m'[39;49;00m)),
                ([33m'[39;49;00m[33mscaler[39;49;00m[33m'[39;49;00m, StandardScaler())
        ])
    
        categorical_transformer = Pipeline(steps=[
                ([33m'[39;49;00m[33mimputer[39;49;00m[33m'[39;49;00m, SimpleImputer(strategy=[33m'[39;49;00m[33mmedian[39;49;00m[33m'[39;49;00m)),
                ([33m'[39;49;00m[33monehot[39;49;00m[33m'[39;49;00m, OneHotEncoder(handle_unknown=[33m'[39;49;00m[33mignore[39;49;00m[33m'[39;49;00m))
            ])
    
        preprocessor = ColumnTransformer(transformers=[
            ([33m'[39;49;00m[33mnum[39;49;00m[33m'[39;49;00m, ordinal_transformer, ordinal_features),
            ([33m'[39;49;00m[33mcat[39;49;00m[33m'[39;49;00m, categorical_transformer, categorical_features)
        ])
    
        preprocessor.fit(data.data)
    
        [34mreturn[39;49;00m  preprocessor
    
    
    [34mdef[39;49;00m [32mtrain_model[39;49;00m(X_train, y_train, preprocessor):
        [33m"""Train model."""[39;49;00m
        [36mprint[39;49;00m([33m"[39;49;00m[33mTraining model.[39;49;00m[33m"[39;49;00m)
        [37m# TODO: ask if we need np.random.seed(...) here[39;49;00m
    
        clf = RandomForestClassifier(n_estimators=[34m50[39;49;00m)
        clf.fit(preprocessor.transform(X_train), y_train)
        [34mreturn[39;49;00m clf
    
    
    [34mdef[39;49;00m [32mserialize_pipeline[39;49;00m(preprocessor, clf):
        [33m"""Serialize preprocessor and model."""[39;49;00m
        [36mprint[39;49;00m([33m"[39;49;00m[33mSerializing preprocessor and model.[39;49;00m[33m"[39;49;00m)
    
        [34mwith[39;49;00m [36mopen[39;49;00m(DATA_DIR + [33m"[39;49;00m[33m/preprocessor.dill[39;49;00m[33m"[39;49;00m, [33m"[39;49;00m[33mwb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m prep_f:
            dill.dump(preprocessor, prep_f)
    
        [34mwith[39;49;00m [36mopen[39;49;00m(DATA_DIR + [33m"[39;49;00m[33m/model.dill[39;49;00m[33m"[39;49;00m, [33m"[39;49;00m[33mwb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m model_f:
            dill.dump(clf, model_f)
    
    
    [34mdef[39;49;00m [32mmain[39;49;00m():
        data, X_train, y_train, X_test, y_test = load_data()
        preprocessor = train_preprocessor(data)
        clf = train_model(X_train, y_train, preprocessor)
    
        serialize_pipeline(preprocessor, clf)
        [34mreturn[39;49;00m preprocessor, clf, data, X_train, y_train, X_test, y_test
    
    
    [34mif[39;49;00m [31m__name__[39;49;00m == [33m"[39;49;00m[33m__main__[39;49;00m[33m"[39;49;00m:
        main()



```python
!python3 train_classifier.py
```

    Loading adult data from alibi.
    Training preprocessor.
    Training model.
    Serializing preprocessor and model.


This script creates two dill-serialized files `preprocess.dill` and `model.dill` that are used by the `Model` class to make prediction:


```python
!pygmentize pipeline/loanclassifier/Model.py
```

    [34mimport[39;49;00m [04m[36mlogging[39;49;00m
    [34mimport[39;49;00m [04m[36mdill[39;49;00m
    [34mimport[39;49;00m [04m[36mos[39;49;00m
    
    
    dirname = os.path.dirname([31m__file__[39;49;00m)
    
    
    [34mclass[39;49;00m [04m[32mModel[39;49;00m:
    
        [34mdef[39;49;00m [32m__init__[39;49;00m([36mself[39;49;00m, *args, **kwargs):
            [33m"""Deserilize preprocessor and model."""[39;49;00m
            [34mwith[39;49;00m [36mopen[39;49;00m(os.path.join(dirname, [33m"[39;49;00m[33mpreprocessor.dill[39;49;00m[33m"[39;49;00m), [33m"[39;49;00m[33mrb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m prep_f:
                [36mself[39;49;00m.preprocessor = dill.load(prep_f)
            [34mwith[39;49;00m [36mopen[39;49;00m(os.path.join(dirname, [33m"[39;49;00m[33mmodel.dill[39;49;00m[33m"[39;49;00m), [33m"[39;49;00m[33mrb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m model_f:
                [36mself[39;49;00m.clf = dill.load(model_f)
    
        [34mdef[39;49;00m [32mpredict[39;49;00m([36mself[39;49;00m, X, feature_names=[]):
            [33m"""Run input X through loanclassifier model."""[39;49;00m
            logging.info([33m"[39;49;00m[33mInput: [39;49;00m[33m"[39;49;00m + [36mstr[39;49;00m(X))
    
            X_prep = [36mself[39;49;00m.preprocessor.transform(X)
            output = [36mself[39;49;00m.clf.predict_proba(X_prep)
    
            logging.info([33m"[39;49;00m[33mOutput: [39;49;00m[33m"[39;49;00m + [36mstr[39;49;00m(output))
            [34mreturn[39;49;00m output


We will in a moment contenrize this Model. You can test how it will work from the notebook:


```python
import sys

sys.path.append("pipeline/loanclassifier")
from Model import Model

model = Model()
```


```python
import numpy as np
import xai
from train_classifier import load_data

data, X_train, y_train, X_test, y_test = load_data()
proba = model.predict(X_test)

pred = np.argmax(proba, axis=1)
xai.metrics_plot(y_test, pred)
```

    Loading adult data from alibi.





<div>
<style scoped>
    .dataframe tbody tr th:only-of-type {
        vertical-align: middle;
    }

    .dataframe tbody tr th {
        vertical-align: top;
    }

    .dataframe thead th {
        text-align: right;
    }
</style>
<table border="1" class="dataframe">
  <thead>
    <tr style="text-align: right;">
      <th></th>
      <th>target</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>precision</th>
      <td>0.704545</td>
    </tr>
    <tr>
      <th>recall</th>
      <td>0.658497</td>
    </tr>
    <tr>
      <th>specificity</th>
      <td>0.913289</td>
    </tr>
    <tr>
      <th>accuracy</th>
      <td>0.852401</td>
    </tr>
    <tr>
      <th>auc</th>
      <td>0.785893</td>
    </tr>
    <tr>
      <th>f1</th>
      <td>0.680743</td>
    </tr>
  </tbody>
</table>
</div>




    
![png](notebook_files/notebook_12_2.png)
    


## Train and test outliers detector

We will now train the outliers detector using another prepared script [train_detector](https://github.com/SeldonIO/seldon-core/blob/master/examples/outliers/alibi-detect-combiner/train_detector.py):


```python
!pygmentize train_detector.py
```

    [34mimport[39;49;00m [04m[36mdill[39;49;00m
    
    [34mimport[39;49;00m [04m[36mnumpy[39;49;00m [34mas[39;49;00m [04m[36mnp[39;49;00m
    
    [34mfrom[39;49;00m [04m[36malibi_detect[39;49;00m[04m[36m.[39;49;00m[04m[36mod[39;49;00m [34mimport[39;49;00m IForest
    [34mfrom[39;49;00m [04m[36malibi_detect[39;49;00m[04m[36m.[39;49;00m[04m[36mutils[39;49;00m[04m[36m.[39;49;00m[04m[36mdata[39;49;00m [34mimport[39;49;00m create_outlier_batch
    [37m# from alibi_detect.utils.saving import save_detector, load_detector[39;49;00m
    [37m# from alibi_detect.utils.visualize import plot_instance_score[39;49;00m
    
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mpreprocessing[39;49;00m [34mimport[39;49;00m StandardScaler
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mimpute[39;49;00m [34mimport[39;49;00m SimpleImputer
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mpipeline[39;49;00m [34mimport[39;49;00m Pipeline
    [34mfrom[39;49;00m [04m[36msklearn[39;49;00m[04m[36m.[39;49;00m[04m[36mcompose[39;49;00m [34mimport[39;49;00m ColumnTransformer
    
    [34mfrom[39;49;00m [04m[36mtrain_classifier[39;49;00m [34mimport[39;49;00m load_data
    
    
    DATA_DIR = [33m"[39;49;00m[33mpipeline/outliersdetector[39;49;00m[33m"[39;49;00m
    
    
    [34mdef[39;49;00m [32mtrain_preprocessor[39;49;00m(data):
        [33m"""Train preprocessor."""[39;49;00m
        [36mprint[39;49;00m([33m"[39;49;00m[33mTraining preprocessor.[39;49;00m[33m"[39;49;00m)
    
        ordinal_features = [
            n [34mfor[39;49;00m (n, _) [35min[39;49;00m [36menumerate[39;49;00m(data.feature_names)
            [34mif[39;49;00m n [35mnot[39;49;00m [35min[39;49;00m data.category_map
        ]
    
        ordinal_transformer = Pipeline(steps=[
                ([33m'[39;49;00m[33mimputer[39;49;00m[33m'[39;49;00m, SimpleImputer(strategy=[33m'[39;49;00m[33mmedian[39;49;00m[33m'[39;49;00m)),
                ([33m'[39;49;00m[33mscaler[39;49;00m[33m'[39;49;00m, StandardScaler())
        ])
    
        preprocessor = ColumnTransformer(transformers=[
            ([33m'[39;49;00m[33mnum[39;49;00m[33m'[39;49;00m, ordinal_transformer, ordinal_features),
        ])
    
        preprocessor.fit(data.data)
    
        [34mreturn[39;49;00m  preprocessor
    
    
    [34mdef[39;49;00m [32mtrain_detector[39;49;00m(data, preprocessor, perc_outlier=[34m5[39;49;00m):
        [33m"""Train outliers detector."""[39;49;00m
    
        [36mprint[39;49;00m([33m"[39;49;00m[33mInitialize outlier detector.[39;49;00m[33m"[39;49;00m)
        od = IForest(threshold=[34mNone[39;49;00m,  n_estimators=[34m100[39;49;00m)
    
        [36mprint[39;49;00m([33m"[39;49;00m[33mTraining on normal data.[39;49;00m[33m"[39;49;00m)
        np.random.seed([34m0[39;49;00m)
        normal_batch = create_outlier_batch(
            data.data, data.target, n_samples=[34m30000[39;49;00m, perc_outlier=[34m0[39;49;00m
        )
    
        X_train = normal_batch.data.astype([33m'[39;49;00m[33mfloat[39;49;00m[33m'[39;49;00m)
        [37m# y_train = normal_batch.target[39;49;00m
    
        od.fit(preprocessor.transform(X_train))
    
        [36mprint[39;49;00m([33m"[39;49;00m[33mTrain on threshold data.[39;49;00m[33m"[39;49;00m)
        np.random.seed([34m0[39;49;00m)
        threshold_batch = create_outlier_batch(
            data.data, data.target, n_samples=[34m1000[39;49;00m, perc_outlier=perc_outlier
        )
        X_threshold = threshold_batch.data.astype([33m'[39;49;00m[33mfloat[39;49;00m[33m'[39;49;00m)
        [37m# y_threshold = threshold_batch.target[39;49;00m
    
        od.infer_threshold(
            preprocessor.transform(X_threshold), threshold_perc=[34m100[39;49;00m - perc_outlier
        )
    
        [34mreturn[39;49;00m od
    
    
    [34mdef[39;49;00m [32mserialize_pipeline[39;49;00m(preprocessor, od):
        [33m"""Serialize preprocessor and model."""[39;49;00m
        [36mprint[39;49;00m([33m"[39;49;00m[33mSerializing preprocessor and model.[39;49;00m[33m"[39;49;00m)
    
        [34mwith[39;49;00m [36mopen[39;49;00m(DATA_DIR + [33m"[39;49;00m[33m/preprocessor.dill[39;49;00m[33m"[39;49;00m, [33m"[39;49;00m[33mwb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m prep_f:
            dill.dump(preprocessor, prep_f)
    
        [34mwith[39;49;00m [36mopen[39;49;00m(DATA_DIR + [33m"[39;49;00m[33m/model.dill[39;49;00m[33m"[39;49;00m, [33m"[39;49;00m[33mwb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m model_f:
            dill.dump(od, model_f)
    
    
    [34mdef[39;49;00m [32mmain[39;49;00m():
        data = load_data()[[34m0[39;49;00m]
        preprocessor = train_preprocessor(data)
        od = train_detector(data, preprocessor)
        serialize_pipeline(preprocessor, od)
    
    
    [34mif[39;49;00m [31m__name__[39;49;00m == [33m"[39;49;00m[33m__main__[39;49;00m[33m"[39;49;00m:
        main()



```python
!python3 train_detector.py
```

    ERROR:fbprophet:Importing plotly failed. Interactive plots will not work.
    Loading adult data from alibi.
    Training preprocessor.
    Initialize outlier detector.
    WARNING:alibi_detect.od.isolationforest:No threshold level set. Need to infer threshold using `infer_threshold`.
    Training on normal data.
    /home/rskolasinski/.local/lib/python3.7/site-packages/sklearn/ensemble/iforest.py:213: FutureWarning: default contamination parameter 0.1 will change in version 0.22 to "auto". This will change the predict method behavior.
      FutureWarning)
    /home/rskolasinski/.local/lib/python3.7/site-packages/sklearn/ensemble/iforest.py:223: FutureWarning: behaviour="old" is deprecated and will be removed in version 0.22. Please use behaviour="new", which makes the decision_function change to match other anomaly detection algorithm API.
      FutureWarning)
    Train on threshold data.
    Serializing preprocessor and model.


In similar fashion to previous Model it will create `dill-serialized` objects used by `Detector` class:


```python
!pygmentize pipeline/outliersdetector/Detector.py
```

    [34mimport[39;49;00m [04m[36mlogging[39;49;00m
    [34mimport[39;49;00m [04m[36mdill[39;49;00m
    [34mimport[39;49;00m [04m[36mos[39;49;00m
    
    [34mimport[39;49;00m [04m[36mnumpy[39;49;00m [34mas[39;49;00m [04m[36mnp[39;49;00m
    
    
    dirname = os.path.dirname([31m__file__[39;49;00m)
    
    
    [34mclass[39;49;00m [04m[32mDetector[39;49;00m:
        [34mdef[39;49;00m [32m__init__[39;49;00m([36mself[39;49;00m, *args, **kwargs):
    
            [34mwith[39;49;00m [36mopen[39;49;00m(os.path.join(dirname, [33m"[39;49;00m[33mpreprocessor.dill[39;49;00m[33m"[39;49;00m), [33m"[39;49;00m[33mrb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m prep_f:
                [36mself[39;49;00m.preprocessor = dill.load(prep_f)
            [34mwith[39;49;00m [36mopen[39;49;00m(os.path.join(dirname, [33m"[39;49;00m[33mmodel.dill[39;49;00m[33m"[39;49;00m), [33m"[39;49;00m[33mrb[39;49;00m[33m"[39;49;00m) [34mas[39;49;00m model_f:
                [36mself[39;49;00m.od = dill.load(model_f)
    
        [34mdef[39;49;00m [32mpredict[39;49;00m([36mself[39;49;00m, X, feature_names=[]):
            logging.info([33m"[39;49;00m[33mInput: [39;49;00m[33m"[39;49;00m + [36mstr[39;49;00m(X))
    
            X_prep = [36mself[39;49;00m.preprocessor.transform(X)
            output = [36mself[39;49;00m.od.predict(X_prep)[[33m'[39;49;00m[33mdata[39;49;00m[33m'[39;49;00m][[33m'[39;49;00m[33mis_outlier[39;49;00m[33m'[39;49;00m]
    
            logging.info([33m"[39;49;00m[33mOutput: [39;49;00m[33m"[39;49;00m + [36mstr[39;49;00m(output))
            [34mreturn[39;49;00m output


You can see how the detector works from this notebook:


```python
import sys

sys.path.append("pipeline/outliersdetector")
from Detector import Detector

detector = Detector()
```

    ERROR:fbprophet:Importing plotly failed. Interactive plots will not work.



```python
import json

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import seaborn as sns

%matplotlib inline

from alibi_detect.utils.data import create_outlier_batch
from sklearn.metrics import confusion_matrix, f1_score

np.random.seed(1)
outlier_batch = create_outlier_batch(
    data.data, data.target, n_samples=1000, perc_outlier=10
)
X_outlier, y_outlier = outlier_batch.data.astype("float"), outlier_batch.target
```


```python
y_pred = detector.predict(X_outlier)
```


```python
labels = outlier_batch.target_names
f1 = f1_score(y_outlier, y_pred)
print("F1 score: {}".format(f1))
cm = confusion_matrix(y_outlier, y_pred)
df_cm = pd.DataFrame(cm, index=labels, columns=labels)
sns.heatmap(df_cm, annot=True, cbar=True, linewidths=0.5)
plt.show()
```

    F1 score: 0.35365853658536583



    
![png](notebook_files/notebook_22_1.png)
    


## Contenerise your models

Before you can deploy classifier `Model` and outliers `Detector` as part of seldon's graph you have to contenerise them.

We will use the s2i to do so with help of provided [Makefile](https://github.com/SeldonIO/seldon-core/blob/master/examples/outliers/alibi-detect-combiner/Makefile):


```python
!pygmentize Makefile
```

    [32m.ONESHELL[39;49;00m:
    
    [32mall[39;49;00m: base loanclassifier outliersdetector combiner
    
    [32mbase[39;49;00m:
    	docker build . -t seldon-core-outliers-base:0.1
    
    [32mloanclassifier[39;49;00m:
    	s2i build pipeline/loanclassifier seldon-core-outliers-base:0.1 loanclassifier:0.1
    
    [32moutliersdetector[39;49;00m:
    	s2i build pipeline/outliersdetector seldon-core-outliers-base:0.1 outliersdetector:0.1
    
    [32mcombiner[39;49;00m:
    	s2i build pipeline/combiner seldon-core-outliers-base:0.1 combiner:0.1



```python
!make
```

or if using Minikube


```python
!eval $(minikube docker-env) && make
```

## Deploy your models separately

Now, you can include your newly build containers as part of seldon deployment.

First, we will create two separate deployments: `loanclassifier` and `outliersdetector`.
Each of them will have their own separate endpoint and can be queried independently depending on your needs. 

#### Deploy separate loanclassifier

![outliers-combiner-1](img/outliers-combiner-1.jpg)


```python
!pygmentize pipeline/loanclassifier.yaml
```

    [94mapiVersion[39;49;00m: machinelearning.seldon.io/v1alpha2
    [94mkind[39;49;00m: SeldonDeployment
    [94mmetadata[39;49;00m:
      [94mlabels[39;49;00m:
        [94mapp[39;49;00m: seldon
      [94mname[39;49;00m: loanclassifier
    [94mspec[39;49;00m:
      [94mname[39;49;00m: loanclassifier
      [94mpredictors[39;49;00m:
      - [94mcomponentSpecs[39;49;00m:
        - [94mspec[39;49;00m:
            [94mcontainers[39;49;00m:
            - [94mimage[39;49;00m: loanclassifier:0.1
              [94mname[39;49;00m: loanclassifier
              [94menv[39;49;00m:
              - [94mname[39;49;00m: SELDON_LOG_LEVEL
                [94mvalue[39;49;00m: DEBUG
        [94mgraph[39;49;00m:
          [94mchildren[39;49;00m: []
          [94mname[39;49;00m: loanclassifier
          [94mtype[39;49;00m: MODEL
          [94mendpoint[39;49;00m:
            [94mtype[39;49;00m: REST
        [94mname[39;49;00m: loanclassifier
        [94mreplicas[39;49;00m: 1



```python
!kubectl apply -f pipeline/loanclassifier.yaml
```

    seldondeployment.machinelearning.seldon.io/loanclassifier created


#### Deploy separate outliers detector

![outliers-combiner-2](img/outliers-combiner-2.jpg)


```python
!pygmentize pipeline/outliersdetector.yaml
```

    [94mapiVersion[39;49;00m: machinelearning.seldon.io/v1alpha2
    [94mkind[39;49;00m: SeldonDeployment
    [94mmetadata[39;49;00m:
      [94mlabels[39;49;00m:
        [94mapp[39;49;00m: seldon
      [94mname[39;49;00m: outliersdetector
    [94mspec[39;49;00m:
      [94mname[39;49;00m: outliersdetector
      [94mpredictors[39;49;00m:
      - [94mcomponentSpecs[39;49;00m:
        - [94mspec[39;49;00m:
            [94mcontainers[39;49;00m:
            - [94mimage[39;49;00m: outliersdetector:0.1
              [94mname[39;49;00m: outliersdetector
              [94menv[39;49;00m:
              - [94mname[39;49;00m: SELDON_LOG_LEVEL
                [94mvalue[39;49;00m: DEBUG
        [94mgraph[39;49;00m:
          [94mchildren[39;49;00m: []
          [94mname[39;49;00m: outliersdetector
          [94mtype[39;49;00m: MODEL
          [94mendpoint[39;49;00m:
            [94mtype[39;49;00m: REST
        [94mname[39;49;00m: outliersdetector
        [94mreplicas[39;49;00m: 1



```python
!kubectl apply -f pipeline/outliersdetector.yaml
```

    seldondeployment.machinelearning.seldon.io/outliersdetector created


#### View newly deployed Kubernetes pods


```python
!kubectl get pods
```

    NAME                                                        READY   STATUS    RESTARTS   AGE
    ambassador-69b784f9d5-b444s                                 1/1     Running   1          22h
    ambassador-69b784f9d5-zkpbv                                 1/1     Running   3          22h
    ambassador-69b784f9d5-zx9w7                                 1/1     Running   3          22h
    loanclassifier-loanclassifier-65e0c2e-7449f4d596-55p8t      2/2     Running   0          2m10s
    outliersdetector-outliersdetector-1a4e53e-b5cd784df-j56tx   2/2     Running   0          2m8s


#### Test deployed components
**IMPORTANT:** If you are using minikube (instead of docker desktop) you have to forward the port first with:
```
kubectl port-forward svc/ambassador 8003:80
```


```python
import json

from seldon_core.seldon_client import SeldonClient
from seldon_core.utils import get_data_from_proto
```


```python
to_explain = X_test[:3]
print(to_explain)
```

    [[  46    5    4    2    8    4    4    0 2036    0   60    9]
     [  52    4    0    2    8    4    2    0    0    0   60    9]
     [  21    4    4    1    2    3    4    1    0    0   20    9]]



```python
sc = SeldonClient(
    gateway="ambassador",
    deployment_name="loanclassifier",
    gateway_endpoint="localhost:8003",
    payload_type="ndarray",
    namespace="seldon",
    transport="rest",
)

prediction = sc.predict(data=to_explain)
get_data_from_proto(prediction.response)
```




    array([[0.9, 0.1],
           [0.9, 0.1],
           [1. , 0. ]])




```python
sc = SeldonClient(
    gateway="ambassador",
    deployment_name="outliersdetector",
    gateway_endpoint="localhost:8003",
    payload_type="ndarray",
    namespace="seldon",
    transport="rest",
)

prediction = sc.predict(data=to_explain)
get_data_from_proto(prediction.response)
```




    array([0., 0., 0.])



## Deploy loanclassifier and outliersdetector with combiner

Another possibility is to use slightly more complicated graph with a `combiner` that will
gather outputs from `loanclassifier` and `outliersdetector`.

Please note that `loanclassifier` and `outliersdetector` are part of `loanclassifier-combined` graph and this deployment is independent from previous two.

In this approach there is a single API endpoint that serves both functionalities.

![outliers-combiner-3](img/outliers-combiner-3.jpg)


```python
!pygmentize pipeline/combiner/Combiner.py
```

    [34mimport[39;49;00m [04m[36mlogging[39;49;00m
    [34mimport[39;49;00m [04m[36mnumpy[39;49;00m [34mas[39;49;00m [04m[36mnp[39;49;00m
    
    [34mclass[39;49;00m [04m[32mCombiner[39;49;00m([36mobject[39;49;00m):
    
        [34mdef[39;49;00m [32maggregate[39;49;00m([36mself[39;49;00m, X, features_names=[]):
            logging.info([33m"[39;49;00m[33mInput: [39;49;00m[33m"[39;49;00m + [36mstr[39;49;00m(X))
            output = {
                [33m"[39;49;00m[33mloanclassifier[39;49;00m[33m"[39;49;00m: X[[34m0[39;49;00m].tolist(),
                [33m"[39;49;00m[33moutliersdetector[39;49;00m[33m"[39;49;00m: X[[34m1[39;49;00m].tolist(),
            }
            logging.info([33m"[39;49;00m[33mOutput: [39;49;00m[33m"[39;49;00m + [36mstr[39;49;00m(output))
            [34mreturn[39;49;00m output



```python
! pygmentize pipeline/combiner.yaml
```

    [94mapiVersion[39;49;00m: machinelearning.seldon.io/v1alpha2
    [94mkind[39;49;00m: SeldonDeployment
    [94mmetadata[39;49;00m:
      [94mlabels[39;49;00m:
        [94mapp[39;49;00m: seldon
      [94mname[39;49;00m: loanclassifier-combined
    [94mspec[39;49;00m:
      [94mannotations[39;49;00m:
        [94mproject_name[39;49;00m: Iris classification
      [94mname[39;49;00m: loanclassifier-combined
      [94mpredictors[39;49;00m:
      - [94mcomponentSpecs[39;49;00m:
        - [94mspec[39;49;00m:
            [94mcontainers[39;49;00m:
            - [94mimage[39;49;00m: loanclassifier:0.1
              [94mname[39;49;00m: loanclassifier
              [94menv[39;49;00m:
              - [94mname[39;49;00m: SELDON_LOG_LEVEL
                [94mvalue[39;49;00m: DEBUG
            - [94mimage[39;49;00m: outliersdetector:0.1
              [94mname[39;49;00m: outliersdetector
              [94menv[39;49;00m:
              - [94mname[39;49;00m: SELDON_LOG_LEVEL
                [94mvalue[39;49;00m: DEBUG
            - [94mimage[39;49;00m: combiner:0.1
              [94mname[39;49;00m: combiner
              [94menv[39;49;00m:
              - [94mname[39;49;00m: SELDON_LOG_LEVEL
                [94mvalue[39;49;00m: DEBUG
        [94mgraph[39;49;00m:
          [94mchildren[39;49;00m:
          - [94mchildren[39;49;00m: []
            [94mname[39;49;00m: loanclassifier
            [94mtype[39;49;00m: MODEL
            [94mendpoint[39;49;00m:
              [94mtype[39;49;00m: REST
          - [94mchildren[39;49;00m: []
            [94mname[39;49;00m: outliersdetector
            [94mtype[39;49;00m: MODEL
            [94mendpoint[39;49;00m:
              [94mtype[39;49;00m: REST
          [94mendpoint[39;49;00m:
            [94mtype[39;49;00m: REST
          [94mname[39;49;00m: combiner
          [94mtype[39;49;00m: COMBINER
        [94mname[39;49;00m: combiner-graph
        [94mreplicas[39;49;00m: 1



```python
!kubectl apply -f pipeline/combiner.yaml
```

    seldondeployment.machinelearning.seldon.io/loanclassifier-combined created



```python
!kubectl get pods
```

    NAME                                                              READY   STATUS    RESTARTS   AGE
    ambassador-69b784f9d5-b444s                                       1/1     Running   1          22h
    ambassador-69b784f9d5-zkpbv                                       1/1     Running   3          22h
    ambassador-69b784f9d5-zx9w7                                       1/1     Running   3          22h
    loanclassifier-combined-combiner-graph-f931ba8-6b8645f8f9-q99nr   4/4     Running   0          28s
    loanclassifier-loanclassifier-65e0c2e-7449f4d596-55p8t            2/2     Running   0          2m50s
    outliersdetector-outliersdetector-1a4e53e-b5cd784df-j56tx         2/2     Running   0          2m48s



```python
sc = SeldonClient(
    gateway="ambassador",
    deployment_name="loanclassifier-combined",
    gateway_endpoint="localhost:8003",
    payload_type="ndarray",
    namespace="seldon",
    transport="rest",
)

prediction = sc.predict(data=to_explain)
output = get_data_from_proto(prediction.response)
```


```python
prediction.response
```




    meta {
      puid: "5ef3jd2oumacs4ig0ldr7ads2a"
      routing {
        key: "combiner"
        value: -1
      }
      requestPath {
        key: "combiner"
        value: "combiner:0.1"
      }
      requestPath {
        key: "loanclassifier"
        value: "loanclassifier:0.1"
      }
      requestPath {
        key: "outliersdetector"
        value: "outliersdetector:0.1"
      }
    }
    jsonData {
      struct_value {
        fields {
          key: "loanclassifier"
          value {
            list_value {
              values {
                list_value {
                  values {
                    number_value: 0.9
                  }
                  values {
                    number_value: 0.1
                  }
                }
              }
              values {
                list_value {
                  values {
                    number_value: 0.9
                  }
                  values {
                    number_value: 0.1
                  }
                }
              }
              values {
                list_value {
                  values {
                    number_value: 1.0
                  }
                  values {
                    number_value: 0.0
                  }
                }
              }
            }
          }
        }
        fields {
          key: "outliersdetector"
          value {
            list_value {
              values {
                number_value: 0.0
              }
              values {
                number_value: 0.0
              }
              values {
                number_value: 0.0
              }
            }
          }
        }
      }
    }




```python
output["loanclassifier"]
```




    [[0.9, 0.1], [0.9, 0.1], [1.0, 0.0]]




```python
output["outliersdetector"]
```




    [0.0, 0.0, 0.0]


