# SKLearn Spacy Reddit Text Classification Example

In this example we will be buiding a text classifier using the reddit content moderation dataset.

For this, we will be using SpaCy for the word tokenization and lemmatization. 

The classification will be done with a Logistic Regression binary classifier.

The steps in this tutorial include:

1) Train and build your NLP model

2) Build your containerized model

3) Test your model as a docker container

4) Run Seldon in your kubernetes cluster

5) Deploy your model with Seldon

6) Interact with your model through API

7) Clean your environment


### Before you start
Make sure you install the following dependencies, as they are critical for this example to work:

* Helm v3.0.0+
* A Kubernetes cluster running v1.13 or above (minkube / docker-for-windows work well if enough RAM)
* kubectl v1.14+
* Python 3.6+
* Python DEV requirements (we'll install them below)

Let's get started! 🚀🔥

## 1) Train and build your NLP model


```python
%%writefile requirements.txt
scikit-learn>=0.23.2
spacy==2.3.2
dill==0.3.2
pandas==1.1.1
```

    Overwriting requirements.txt



```python
!pip install -r requirements.txt
```


```python
import pandas as pd 
from sklearn.model_selection import train_test_split
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from seldon_core.seldon_client import SeldonClient
import dill
import sys, os

# This import may take a while as it will download the Spacy ENGLISH model
from ml_utils import CleanTextTransformer, SpacyTokenTransformer
```


```python
df_cols = ["prev_idx", "parent_idx", "body", "removed"]

TEXT_COLUMN = "body" 
CLEAN_COLUMN = "clean_body"
TOKEN_COLUMN = "token_body"

# Downloading the 50k reddit dataset of moderated comments
df = pd.read_csv("https://raw.githubusercontent.com/axsauze/reddit-classification-exploration/master/data/reddit_train.csv", 
                         names=df_cols, skiprows=1, encoding="ISO-8859-1")

df.head()
```




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
      <th>prev_idx</th>
      <th>parent_idx</th>
      <th>body</th>
      <th>removed</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>0</th>
      <td>8756</td>
      <td>8877</td>
      <td>Always be wary of news articles that cite unpu...</td>
      <td>0</td>
    </tr>
    <tr>
      <th>1</th>
      <td>7330</td>
      <td>7432</td>
      <td>The problem I have with this is that the artic...</td>
      <td>0</td>
    </tr>
    <tr>
      <th>2</th>
      <td>15711</td>
      <td>15944</td>
      <td>This is indicative of a typical power law, and...</td>
      <td>0</td>
    </tr>
    <tr>
      <th>3</th>
      <td>1604</td>
      <td>1625</td>
      <td>This doesn't make sense. Chess obviously trans...</td>
      <td>0</td>
    </tr>
    <tr>
      <th>4</th>
      <td>13327</td>
      <td>13520</td>
      <td>1. I dispute that gene engineering is burdenso...</td>
      <td>0</td>
    </tr>
  </tbody>
</table>
</div>




```python
# Let's see how many examples we have of each class
df["removed"].value_counts().plot.bar()
```




    <matplotlib.axes._subplots.AxesSubplot at 0x7f9db626d1d0>




![png](sklearn_spacy_text_classifier_example_files/sklearn_spacy_text_classifier_example_5_1.png)



```python
x = df["body"].values
y = df["removed"].values
x_train, x_test, y_train, y_test = train_test_split(
    x, y, 
    stratify=y, 
    random_state=42, 
    test_size=0.1, shuffle=True)
```


```python
# Clean the text
clean_text_transformer = CleanTextTransformer()
x_train_clean = clean_text_transformer.transform(x_train)
```


```python
# Tokenize the text and get the lemmas
spacy_tokenizer = SpacyTokenTransformer()
x_train_tokenized = spacy_tokenizer.transform(x_train_clean)
```


```python
# Build tfidf vectorizer
tfidf_vectorizer = TfidfVectorizer(
    max_features=10000,
    preprocessor=lambda x: x, 
    tokenizer=lambda x: x, 
    token_pattern=None,
    ngram_range=(1, 3))

tfidf_vectorizer.fit(x_train_tokenized)
```




    TfidfVectorizer(max_features=10000, ngram_range=(1, 3),
                    preprocessor=<function <lambda> at 0x7f9db31f58c0>,
                    token_pattern=None,
                    tokenizer=<function <lambda> at 0x7f9db31f5830>)




```python
# Transform our tokens to tfidf vectors
x_train_tfidf = tfidf_vectorizer.transform(x_train_tokenized)
```


```python
# Train logistic regression classifier
lr = LogisticRegression(C=0.1, solver='sag')
lr.fit(x_train_tfidf, y_train)
```




    LogisticRegression(C=0.1, solver='sag')




```python
# These are the models we'll deploy
with open('tfidf_vectorizer.model', 'wb') as model_file:
    dill.dump(tfidf_vectorizer, model_file)
with open('lr.model', 'wb') as model_file:
    dill.dump(lr, model_file)
```

## 2) Build your containerized model


```python
%%writefile RedditClassifier.py
import dill

from ml_utils import CleanTextTransformer, SpacyTokenTransformer


class RedditClassifier(object):
    def __init__(self):

        self._clean_text_transformer = CleanTextTransformer()
        self._spacy_tokenizer = SpacyTokenTransformer()

        with open("tfidf_vectorizer.model", "rb") as model_file:
            self._tfidf_vectorizer = dill.load(model_file)

        with open("lr.model", "rb") as model_file:
            self._lr_model = dill.load(model_file)

    def predict(self, X, feature_names):
        clean_text = self._clean_text_transformer.transform(X)
        spacy_tokens = self._spacy_tokenizer.transform(clean_text)
        tfidf_features = self._tfidf_vectorizer.transform(spacy_tokens)
        predictions = self._lr_model.predict_proba(tfidf_features)
        return predictions
```

    Overwriting RedditClassifier.py



```python
# test that our model works
from RedditClassifier import RedditClassifier
# With one sample
sample = x_test[0:1]
print(sample)
print(RedditClassifier().predict(sample, ["feature_name"]))
```

    ['This is the study that the article is based on:\r\n\r\nhttps://www.nature.com/articles/nature25778.epdf']
    [[0.82791777 0.17208223]]


### Create Docker Image with the S2i utility
Using the S2I command line interface we wrap our current model to seve it through the Seldon interface


```python
# To create a docker image we need to create the .s2i folder configuration as below:
!cat .s2i/environment
```

    MODEL_NAME=RedditClassifier
    API_TYPE=REST
    SERVICE_TYPE=MODEL
    PERSISTENCE=0



```python
# As well as a requirements.txt file with all the relevant dependencies
!cat requirements.txt
```

    scikit-learn>=0.23.2
    spacy==2.3.2
    dill==0.3.2
    pandas==1.1.1



```python
%%writefile
FROM seldonio/seldon-core-s2i-python3:1.3.0-dev

RUN pip install spacy
RUN python -m spacy download en_core_web_sm
```

    UsageError: the following arguments are required: filename



```bash
%%bash
docker build . -t seldonio/seldon-core-spacy-base:0.1
s2i build . seldonio/seldon-core-spacy-base:0.1 seldonio/reddit-classifier:0.1
```

## 3) Test your model as a docker container


```python
# Remove previously deployed containers for this model
!docker rm -f reddit_predictor
```

    Error: No such container: reddit_predictor



```python
!docker run --name "reddit_predictor" -d --rm -p 5001:5000 seldonio/reddit-classifier:0.1
```

    2743a0561c99be7371dbbbb6c87a036c8fe22690bcba6da03d8041c1011546cc



```python
# We now test the REST endpoint expecting the same result
endpoint = "0.0.0.0:5001"
batch = sample
payload_type = "ndarray"

sc = SeldonClient(microservice_endpoint=endpoint)
response = sc.microservice(
    data=batch,
    method="predict",
    payload_type=payload_type,
    names=["tfidf"])

print(response)
```

    Success:True message:
    Request:
    data {
      names: "tfidf"
      ndarray {
        values {
          string_value: "This is the study that the article is based on:\r\n\r\nhttps://www.nature.com/articles/nature25778.epdf"
        }
      }
    }
    
    Response:
    meta {
    }
    data {
      names: "t:0"
      names: "t:1"
      ndarray {
        values {
          list_value {
            values {
              number_value: 0.8285423647440985
            }
            values {
              number_value: 0.17145763525590152
            }
          }
        }
      }
    }
    



```python
# We now stop it to run it in docker
!docker stop reddit_predictor
```

    reddit_predictor


## 4) Run Seldon in your kubernetes cluster


## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

## 5) Deploy your model with Seldon
We can now deploy our model by using the Seldon graph definition:


```python
%%writefile reddit_clf.json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "reddit-classifier"
    },
    "spec": {
        "annotations": {
            "project_name": "Reddit classifier",
            "deployment_version": "v1"
        },
        "name": "reddit-classifier",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/reddit-classifier:0.1",
                                "imagePullPolicy": "IfNotPresent",
                                "name": "classifier",
                                "resources": {
                                    "requests": {
                                        "memory": "1Mi"
                                    }
                                }
                            }
                        ],
                        "terminationGracePeriodSeconds": 20
                    }
                }],
                "graph": {
                    "children": [],
                    "name": "classifier",
                    "endpoint": {
            "type" : "REST"
            },
                    "type": "MODEL"
                },
                "name": "single-model",
                "replicas": 1,
        "annotations": {
            "predictor_version" : "v1"
        }
            }
        ]
    }
}
```

    Overwriting reddit_clf.json


Note: if you are using kind preload image first with
```bash
kind load docker-image reddit-classifier:0.1 --name <name of your cluster>
```


```python
!kubectl apply -f reddit_clf.json
```

    seldondeployment.machinelearning.seldon.io/reddit-classifier created



```python
!kubectl get pods 
```

    NAME                                                           READY   STATUS    RESTARTS   AGE
    reddit-classifier-single-model-0-classifier-6fb8dbfd87-w8stj   2/2     Running   0          27s


## 6) Interact with your model through API
Now that our Seldon Deployment is live, we are able to interact with it through its API.

There are two options in which we can interact with our new model. These are:

a) Using CURL from the CLI (or another rest client like Postman)

b) Using the Python SeldonClient

#### a) Using CURL from the CLI


```bash
%%bash
curl -s -H 'Content-Type: application/json' \
    -d '{"data": {"names": ["text"], "ndarray": ["Hello world this is a test"]}}' \
    http://localhost:8003/seldon/seldon/reddit-classifier/api/v1.0/predictions
```

    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6821638979867455,0.3178361020132546]]},"meta":{}}


#### b) Using the Python SeldonClient


```python
from seldon_core.seldon_client import SeldonClient
import numpy as np

sc = SeldonClient(
    gateway="ambassador", 
    transport="rest",
    gateway_endpoint="localhost:8003",   # Make sure you use the port above
    namespace="seldon"
)

client_prediction = sc.predict(
    data=np.array(["Hello world this is a test"]), 
    deployment_name="reddit-classifier",
    names=["text"],
    payload_type="ndarray",
)

print(client_prediction)
```

    Success:True message:
    Request:
    meta {
    }
    data {
      names: "text"
      ndarray {
        values {
          string_value: "Hello world this is a test"
        }
      }
    }
    
    Response:
    {'data': {'names': ['t:0', 't:1'], 'ndarray': [[0.6821638979867455, 0.3178361020132546]]}, 'meta': {}}


## 7) Clean your environment


```python
!kubectl delete -f reddit_clf.json
```

    seldondeployment.machinelearning.seldon.io "reddit-classifier" deleted

