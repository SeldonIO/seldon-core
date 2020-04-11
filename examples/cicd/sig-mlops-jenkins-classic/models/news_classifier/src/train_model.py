from sklearn.datasets import fetch_20newsgroups
from sklearn.pipeline import Pipeline
from sklearn.feature_extraction.text import TfidfTransformer, CountVectorizer
from sklearn.naive_bayes import MultinomialNB
import numpy as np
import joblib


def fetch_data():
    categories = ["alt.atheism", "soc.religion.christian", "comp.graphics", "sci.med"]
    twenty_train = fetch_20newsgroups(
        subset="train", categories=categories, shuffle=True, random_state=42
    )
    twenty_test = fetch_20newsgroups(
        subset="test", categories=categories, shuffle=True, random_state=42
    )

    return twenty_train, twenty_test


def build_train_model(twenty_train):
    text_clf = Pipeline(
        [
            ("vect", CountVectorizer()),
            ("tfidf", TfidfTransformer()),
            ("clf", MultinomialNB()),
        ]
    )

    text_clf.fit(twenty_train.data, twenty_train.target)
    return text_clf


def print_accuracy(twenty_test, text_clf):
    predicted = text_clf.predict(twenty_test.data)
    print(f"Accuracy: {np.mean(predicted == twenty_test.target):.2f}")


def save_model(text_clf):
    joblib.dump(text_clf, "src/model.joblib")


if __name__ == "__main__":
    twenty_train, twenty_test = fetch_data()
    text_clf = build_train_model(twenty_train)
    print_accuracy(twenty_test, text_clf)
    save_model(text_clf)
