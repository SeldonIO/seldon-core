# Scikit-Learn Server



```python
import os

import joblib
import numpy as np
from sklearn import datasets
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline


def main():
    clf = LogisticRegression()
    p = Pipeline([("clf", clf)])
    print("Training model...")
    p.fit(X, y)
    print("Model trained!")

    filename_p = "model.joblib"
    print("Saving model in %s" % filename_p)
    joblib.dump(p, filename_p)
    print("Model saved!")


if __name__ == "__main__":
    print("Loading iris data set...")
    iris = datasets.load_iris()
    X, y = iris.data, iris.target
    print("Dataset loaded!")
    main()
```

    Loading iris data set...
    Dataset loaded!
    Training model...
    Model trained!
    Saving model in model.joblib
    Model saved!


    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/sklearn/linear_model/logistic.py:432: FutureWarning: Default solver will be changed to 'lbfgs' in 0.22. Specify a solver to silence this warning.
      FutureWarning)
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/sklearn/linear_model/logistic.py:469: FutureWarning: Default multi_class will be changed to 'auto' in 0.22. Specify the multi_class option to silence this warning.
      "this warning.", FutureWarning)


Wrap model using s2i

## REST test


```python
!cd .. && make build_rest
```


```python
!docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"file:///model"}]' -v ${PWD}:/model seldonio/sklearnserver_rest:0.1
```

    85ebfc6c41ef145b578077809af81a23ecb6c7ffe261645b098466d6fcda6ecb


Send some random features that conform to the contract


```python
!seldon-core-tester contract.json 0.0.0.0 5000 -p
```

    ----------------------------------------
    SENDING NEW REQUEST:
    
    [[6.834 4.605 7.238 2.832]]
    RECEIVED RESPONSE:
    meta {
    }
    data {
      names: "t:0"
      names: "t:1"
      names: "t:2"
      ndarray {
        values {
          list_value {
            values {
              number_value: 7.698570018103115e-05
            }
            values {
              number_value: 0.037101590872860316
            }
            values {
              number_value: 0.9628214234269586
            }
          }
        }
      }
    }
    
    



```python
!docker rm sklearnserver --force
```

    sklearnserver



```python
!docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"method","value":"predict"},{"type":"STRING","name":"model_uri","value":"file:///model"}]' -v ${PWD}:/model seldonio/sklearnserver_rest:0.1
```

    d7298dbeaee7508c995d817901b84cf983397003cd1eb74dabc46fd14dad49b0



```python
!seldon-core-tester contract.json 0.0.0.0 5000 -p
```

    ----------------------------------------
    SENDING NEW REQUEST:
    
    [[7.22  3.214 1.305 2.948]]
    RECEIVED RESPONSE:
    meta {
    }
    data {
      ndarray {
        values {
          number_value: 0.0
        }
      }
    }
    
    



```python
!docker rm sklearnserver --force
```

    sklearnserver


## grpc test


```python
!cd .. && make build_grpc
```


```python
!docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"file:///model"}]' -v ${PWD}:/model seldonio/sklearnserver_grpc:0.1
```

    9d0218b348e186596717736035bf67fc75f91ec0bdf8152b9d1ad9734d842d54


Test using NDArray payload


```python
!seldon-core-tester contract.json 0.0.0.0 5000 -p --grpc
```

    ----------------------------------------
    SENDING NEW REQUEST:
    
    [[6.538 4.217 6.519 0.217]]
    RECEIVED RESPONSE:
    meta {
    }
    data {
      names: "t:0"
      names: "t:1"
      names: "t:2"
      ndarray {
        values {
          list_value {
            values {
              number_value: 0.003966041860793068
            }
            values {
              number_value: 0.8586797745038719
            }
            values {
              number_value: 0.13735418363533516
            }
          }
        }
      }
    }
    
    


Test using Tensor payload


```python
!seldon-core-tester contract.json 0.0.0.0 5000 -p --grpc --tensor
```

    ----------------------------------------
    SENDING NEW REQUEST:
    
    [[4.404 4.341 5.101 0.219]]
    RECEIVED RESPONSE:
    meta {
    }
    data {
      names: "t:0"
      names: "t:1"
      names: "t:2"
      tensor {
        shape: 1
        shape: 3
        values: 0.10494571335925532
        values: 0.6017695103262425
        values: 0.29328477631450234
      }
    }
    
    



```python
!docker rm sklearnserver --force
```

    sklearnserver



```python
def x(a=None, b=2):
    print(a, b)
```


```python
x(b=3, a=1)
```

    1 3



```python

```
