import numpy as np
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score
from sklearn.model_selection import train_test_split
from sklearn.pipeline import Pipeline
from typing import Tuple, Union
import requests
from requests import RequestException
from typing import Dict
import joblib
import tarfile
from io import BytesIO


MOVIESENTIMENT_URLS = [
    "https://storage.googleapis.com/seldon-datasets/sentence_polarity_v1/rt-polaritydata.tar.gz",
    "http://www.cs.cornell.edu/People/pabo/movie-review-data/rt-polaritydata.tar.gz",
]


# copy from alibi project to avoid dependency for model training
def fetch_movie_sentiment(
    return_X_y: bool = False, url_id: int = 0
) -> Union[Dict, Tuple[list, list]]:
    """
    The movie review dataset, equally split between negative and positive reviews.
    Parameters
    ----------
    return_X_y
        If true, return features X and labels y as Python lists, if False return a Bunch object
    url_id
        Index specifying which URL to use for downloading
    Returns
    -------
    Dict
        Movie reviews and sentiment labels (0 means 'negative' and 1 means 'positive').
    (data, target)
        Tuple if ``return_X_y`` is true
    """
    url = MOVIESENTIMENT_URLS[url_id]
    try:
        resp = requests.get(url, timeout=2)
        resp.raise_for_status()
    except RequestException:
        print("Could not connect, URL may be out of service")
        raise

    tar = tarfile.open(fileobj=BytesIO(resp.content), mode="r:gz")
    data = []
    labels = []
    for i, member in enumerate(tar.getnames()[1:]):
        f = tar.extractfile(member)
        for line in f.readlines():
            try:
                line.decode("utf8")
            except UnicodeDecodeError:
                continue
            data.append(line.decode("utf8").strip())
            labels.append(i)
    tar.close()
    if return_X_y:
        return data, labels

    target_names = ["negative", "positive"]
    return dict(data=data, target=labels, target_names=target_names)


# load data
movies = fetch_movie_sentiment()
movies.keys()
data = movies["data"]
labels = movies["target"]
target_names = movies["target_names"]

# define train and test set
np.random.seed(0)
train, test, train_labels, test_labels = train_test_split(
    data, labels, test_size=0.2, random_state=42
)
train, val, train_labels, val_labels = train_test_split(
    train, train_labels, test_size=0.1, random_state=42
)
train_labels = np.array(train_labels)
test_labels = np.array(test_labels)
val_labels = np.array(val_labels)

# define and  train an cnn model
vectorizer = CountVectorizer(min_df=1)
clf = LogisticRegression(solver="liblinear")
pipeline = Pipeline([("preprocess", vectorizer), ("clf", clf)])

print("Training ...")
pipeline.fit(train, train_labels)
print("Training done!")

preds_train = pipeline.predict(train)
preds_val = pipeline.predict(val)
preds_test = pipeline.predict(test)
print("Train accuracy", accuracy_score(train_labels, preds_train))
print("Validation accuracy", accuracy_score(val_labels, preds_val))
print("Test accuracy", accuracy_score(test_labels, preds_test))

print("Saving Model to model.joblib")
joblib.dump(pipeline, "model.joblib")
