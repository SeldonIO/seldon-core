# Isolation Forest (IF) Algorithm Documentation

The aim of this document is to explain the Isolation Forest algorithm in Seldon's outlier detection framework.

First, we provide a high level overview of the algorithm and the use case, then we will give a detailed explanation of the implementation.

## Overview

Outlier detection has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. The available data is typically unlabeled and detection needs to be done in real-time. The outlier detector can be used as a standalone algorithm, or to detect anomalies in the input data of another predictive model.

The IF outlier detection algorithm predicts whether the input features are an outlier or not, dependent on a threshold level set by the user. The algorithm needs to be pretrained first on a representable batch of data.

As observations arrive, the algorithm will:
- calculate an anomaly score for the observation
- predict that the observation is an outlier if the anomaly score is below the threshold level

## Why Isolation Forests?

Isolation forests are tree based models specifically used for outlier detection. The IF isolates observations by randomly selecting a feature and then randomly selecting a split value between the maximum and minimum values of the selected feature. The number of splittings required to isolate a sample is equivalent to the path length from the root node to the terminating node. This path length, averaged over a forest of random trees, is a measure of normality and is used to define an anomaly score. Outliers can typically be isolated quicker, leading to shorter paths. In the scikit-learn implementation, lower anomaly scores indicate that the probability of an observation being an outlier is higher.

## Implementation

### 1. Defining and training the IF model

The model takes 4 hyperparameters:

- contamination: the fraction of expected outliers in the data set
- number of estimators: the number of base estimators; number of trees in the forest
- max samples: fraction of samples used for each base estimator
- max features: fraction of features used for each base estimator

``` python
!python train.py \
--dataset 'kddcup99' \
--samples 50000 \
--keep_cols "$cols_str" \
--contamination .1 \
--n_estimators 100 \
--max_samples .8 \
--max_features 1. \
--save_path './models/'
```

The model is saved in the folder specified by "save_path".

### 2. Making predictions

In order to make predictions, which can then be served by Seldon Core, the pre-trained model is loaded when defining an OutlierIsolationForest object. The "threshold" argument defines below which anomaly score a sample is classified as an outlier. The threshold is a key hyperparameter and needs to be picked carefully for each application. The OutlierIsolationForest class inherits from the CoreIsolationForest class in ```CoreIsolationForest.py```.

``` python
class CoreIsolationForest(object):
    """ Outlier detection using Isolation Forests.
    
    Parameters
    ----------
        threshold (float) : anomaly score threshold; scores below threshold are outliers
     
    Functions
    ----------
        predict : detect and return outliers
        transform_input : detect outliers and return input features
        send_feedback : add target labels as part of the feedback loop
        tags : add metadata for input transformer
        metrics : return custom metrics
    """
    
    def __init__(self,threshold=0.,load_path='./models/'):
        
        logger.info("Initializing model")
        self.threshold = threshold
        self.N = 0 # total sample count up until now
        self.nb_outliers = 0
        
        # load pre-trained model
        with open(load_path + 'model.pickle', 'rb') as f:
            self.clf = pickle.load(f)
```

```python
class OutlierIsolationForest(CoreIsolationForest):
    """ Outlier detection using Isolation Forests.
    
    Parameters
    ----------
        threshold (float) : anomaly score threshold; scores below threshold are outliers
     
    Functions
    ----------
        send_feedback : add target labels as part of the feedback loop
        metrics : return custom metrics
    """
    def __init__(self,threshold=0.,load_path='./models/'):
        
        super().__init__(threshold=threshold, load_path=load_path)
```

The actual outlier detection is done by the ```_get_preds``` method which is invoked by ```predict``` or ```transform_input``` dependent on whether the detector is defined as respectively a model or a transformer.

``` python
def predict(self, X, feature_names):
    """ Return outlier predictions.

    Parameters
    ----------
    X : array-like
    feature_names : array of feature names (optional)
    """
    logger.info("Using component as a model")
    return self._get_preds(X)
```

```python
def transform_input(self, X, feature_names):
    """ Transform the input. 
    Used when the outlier detector sits on top of another model.

    Parameters
    ----------
    X : array-like
    feature_names : array of feature names (optional)
    """
    logger.info("Using component as an outlier-detector transformer")
    self.prediction_meta = self._get_preds(X)
    return X
```

## References

Scikit-learn Isolation Forest:
- https://scikit-learn.org/stable/modules/generated/sklearn.ensemble.IsolationForest.html