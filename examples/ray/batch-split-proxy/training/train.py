import argparse
import math
import pandas as pd
from simpletransformers.model import TransformerModel
from sklearn.model_selection import train_test_split

from alibi.datasets import fetch_movie_sentiment


def prepare_data(test_size):
    # load data
    X, y = fetch_movie_sentiment(return_X_y=True)

    # prepare data
    data = pd.DataFrame()
    data["text"] = X
    data["labels"] = y

    if math.isclose(test_size, 0.0):
        return data, None
    else:
        train, test = train_test_split(data, test_size=test_size)
        return train, test


def run(args):
    train, test = prepare_data(args.eval)
    model = TransformerModel(
        "roberta", "roberta-base", args=({"fp16": False}), use_cuda=False
    )
    model.train_model(train)

    if test is not None:
        result, model_outputs, wrong_predictions = model.eval_model(test)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Train a RoBERTa movie sentiment classifier"
    )
    parser.add_argument(
        "--eval",
        type=float,
        default=0.0,
        help="Proportion of dataset to set aside for evaluation",
    )
    args = parser.parse_args()

    run(args)
