import click
import numpy as np
import dill
from sklearn.feature_extraction.text import TfidfVectorizer

@click.command()
@click.option('--in-path', default="/mnt/tokenized_text.data")
@click.option('--out-path', default="/mnt/tfidf_vectors.data")
@click.option('--max-features', default=1000)
@click.option('--ngram-range', default=3)
@click.option('--action', default="predict", 
        type=click.Choice(['predict', 'train']))
@click.option('--model-path', default="/mnt/tfidf.model")
def run_pipeline(
        in_path, 
        out_path, 
        max_features,
        ngram_range,
        action,
        model_path):

    with open(in_path, 'rb') as in_f:
        x = dill.load(in_f)

    if action == "train":
        tfidf_vectorizer = TfidfVectorizer(
            max_features=max_features,
            preprocessor=lambda x: x, # We're using cleantext
            tokenizer=lambda x: x, # We're using spacy
            token_pattern=None,
            ngram_range=(1, ngram_range))

        tfidf_vectorizer.fit(x)
        with open(model_path, "wb") as model_f:
            dill.dump(tfidf_vectorizer, model_f)

    elif action == "predict":
        with open(model_path, "rb") as model_f:
            tfidf_vectorizer = dill.load(model_f)

    y = tfidf_vectorizer.transform(x)

    with open(out_path, "wb") as out_f:
        dill.dump(y, out_f)

if __name__ == "__main__":
    run_pipeline()

