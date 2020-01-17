import mlflow

from argparse import ArgumentParser

from pyspark.ml.regression import LinearRegression
from pyspark.ml.feature import VectorAssembler
from pyspark.ml import Pipeline

from ..common import eval_metrics, read_data

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
    "--l1-ratio",
    action="store",
    dest="l1_ratio",
    type=float,
    default=0.2,
    help="L1 regularizer ratio.",
)


def train(alpha, l1_ratio):
    # Split data into training and test datasets.
    data = read_data()
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
