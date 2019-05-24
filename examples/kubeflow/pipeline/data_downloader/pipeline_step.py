import click
import numpy as np
import dill
import pandas as pd

@click.command()
@click.option('--labels-path', default="/mnt/labels.data")
@click.option('--features-path', default="/mnt/features.data")
@click.option('--csv-url', default="https://raw.githubusercontent.com/axsauze/reddit-classification-exploration/master/data/reddit_train.csv")
@click.option('--csv-encoding', default="ISO-8859-1")
@click.option('--features-column', default="BODY")
@click.option('--labels-column', default="REMOVED")
def run_pipeline(
        labels_path, 
        features_path,
        csv_url, 
        csv_encoding,
        features_column,
        labels_column):

    df = pd.read_csv(csv_url, encoding=csv_encoding)

    x = df[features_column].values

    with open(features_path, "wb") as out_f:
        dill.dump(x, out_f)

    y = df[labels_column].values

    with open(labels_path, "wb") as out_f:
        dill.dump(y, out_f)

if __name__ == "__main__":
    run_pipeline()

