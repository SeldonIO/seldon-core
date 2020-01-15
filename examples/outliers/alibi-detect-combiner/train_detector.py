import dill

import numpy as np

from alibi_detect.od import IForest
from alibi_detect.utils.data import create_outlier_batch
# from alibi_detect.utils.saving import save_detector, load_detector
# from alibi_detect.utils.visualize import plot_instance_score

from sklearn.preprocessing import StandardScaler
from sklearn.impute import SimpleImputer
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer

from train_classifier import load_data


DATA_DIR = "pipeline/outliersdetector"


def train_preprocessor(data):
    """Train preprocessor."""
    print("Training preprocessor.")

    ordinal_features = [
        n for (n, _) in enumerate(data.feature_names)
        if n not in data.category_map
    ]

    ordinal_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='median')),
            ('scaler', StandardScaler())
    ])

    preprocessor = ColumnTransformer(transformers=[
        ('num', ordinal_transformer, ordinal_features),
    ])

    preprocessor.fit(data.data)

    return  preprocessor


def train_detector(data, preprocessor, perc_outlier=5):
    """Train outliers detector."""

    print("Initialize outlier detector.")
    od = IForest(threshold=None,  n_estimators=100)

    print("Training on normal data.")
    np.random.seed(0)
    normal_batch = create_outlier_batch(
        data.data, data.target, n_samples=30000, perc_outlier=0
    )

    X_train = normal_batch.data.astype('float')
    # y_train = normal_batch.target

    od.fit(preprocessor.transform(X_train))

    print("Train on threshold data.")
    np.random.seed(0)
    threshold_batch = create_outlier_batch(
        data.data, data.target, n_samples=1000, perc_outlier=perc_outlier
    )
    X_threshold = threshold_batch.data.astype('float')
    # y_threshold = threshold_batch.target

    od.infer_threshold(
        preprocessor.transform(X_threshold), threshold_perc=100 - perc_outlier
    )

    return od


def serialize_pipeline(preprocessor, od):
    """Serialize preprocessor and model."""
    print("Serializing preprocessor and model.")

    with open(DATA_DIR + "/preprocessor.dill", "wb") as prep_f:
        dill.dump(preprocessor, prep_f)

    with open(DATA_DIR + "/model.dill", "wb") as model_f:
        dill.dump(od, model_f)


def main():
    data = load_data()[0]
    preprocessor = train_preprocessor(data)
    od = train_detector(data, preprocessor)
    serialize_pipeline(preprocessor, od)


if __name__ == "__main__":
    main()
