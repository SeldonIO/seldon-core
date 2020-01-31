import mlflow
import numpy as np

from argparse import ArgumentParser

from mlflow import spark as mlflow_spark
from pyspark import SparkContext
from pyspark.sql import SparkSession
from pyspark.ml.regression import LinearRegression
from pyspark.ml.feature import VectorAssembler
from pyspark.ml import Pipeline

from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score


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


def read_data(spark):
    data = spark.read.option("header", "true").csv(
        "../wine-quality.csv", inferSchema=True
    )
    pdf = data.toPandas()

    # We normalize the inputs to both the SparkML & TensorFlow models
    # so that they have the same input schema.
    for col in pdf.columns[:-1]:
        pdf[col] = (pdf[col] - pdf[col].mean()) / pdf[col].std()

    return data


def train(alpha, l1_ratio):
    # Split data into training and test datasets.
    spark = SparkSession.builder.getOrCreate()
    data = read_data(spark)
    (training, test) = data.randomSplit([0.8, 0.2])

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
        mlflow_spark.log_model(lrModel, "")

        # Evaluate the model on the test set.
        predictions = lrModel.transform(test)
        predict_df = predictions.toPandas()
        (rmse, mae, r2) = eval_metrics(predict_df["quality"], predict_df["prediction"])
        mlflow.log_metric("rmse", rmse)
        mlflow.log_metric("r2", r2)


if __name__ == "__main__":
    args = parser.parse_args()
    sc = SparkContext()

    train(args.alpha, args.l1_ratio)

    sc.stop()
