import mlflow
import pandas as pd
import numpy as np

from argparse import ArgumentParser

from pyspark.ml.regression import LinearRegression
from pyspark.ml.feature import VectorAssembler
from pyspark.ml import Pipeline

from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score
from sklearn.model_selection import train_test_split


parser = ArgumentParser()
parser.add_argument(
    "-a",
    "--alpha",
    action="store",
    dest="alpha",
    type=float,
    default=0.5,
    help="Alpha coefficient.",
)
parser.add_argument(
    "-l",
    "--l1_ratio",
    action="store",
    dest="l1_ratio",
    type=float,
    default=0.2,
    help="L1 regularizer ratio.",
)


def eval_metrics(actual, pred):
    rmse = np.sqrt(mean_squared_error(actual, pred))
    mae = mean_absolute_error(actual, pred)
    r2 = r2_score(actual, pred)
    return rmse, mae, r2


def read_data():
    data = pd.read_csv("../wine-quality.csv")
    data.head()

    # We normalize the inputs to both the SparkML & TensorFlow models
    # so that they have the same input schema.
    for col in data.columns[:-1]:
        data[col] = (data[col] - data[col].mean()) / data[col].std()

    return data


def train(alpha, l1_ratio):
    # Split data into training and test datasets.
    data = read_data()
    (training, test) = train_test_split(data, train_size=0.8)

    # Assemble feature columns into a vector (excluding the "quality" label).
    assembler = VectorAssembler(inputCols=data.columns[0:-1], outputCol="features")

    # Create elastic net regressor based on alpha  & l1 ratio hyperparameters.
    lr = LinearRegression(
        maxIter=10, regParam=alpha, elasticNetParam=l1_ratio, labelCol="quality"
    )

    # Create SparkML pipeline, which we can save as one combined model.
    pipeline = Pipeline(stages=[assembler, lr])

    with mlflow.start_run(run_name="spark-a%s-l%s" % (alpha, l1_ratio)):
        mlflow.log_param("alpha", alpha)
        mlflow.log_param("l1_ratio", l1_ratio)

        # Train and save the model.
        lrModel = pipeline.fit(training)
        mlflow.spark.log_model(lrModel, "")

        # Evaluate the model on the test set.
        predictions = lrModel.transform(test)
        predict_df = predictions.toPandas()
        (rmse, mae, r2) = eval_metrics(predict_df["quality"], predict_df["prediction"])
        mlflow.log_metric("rmse", rmse)
        mlflow.log_metric("r2", r2)


if __name__ == "__main__":
    args = parser.parse_args()
    train(args.alpha, args.l1_ratio)
