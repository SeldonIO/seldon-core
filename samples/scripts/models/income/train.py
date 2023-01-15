import joblib
import numpy as np
from alibi.datasets import fetch_adult
from sklearn.compose import ColumnTransformer
from sklearn.ensemble import RandomForestClassifier
from sklearn.impute import SimpleImputer
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, StandardScaler


def main():
    np.random.seed(0)

    # prepare data
    adult = fetch_adult()
    data = adult.data
    target = adult.target
    feature_names = adult.feature_names
    category_map = adult.category_map

    data_perm = np.random.permutation(np.c_[data, target])
    data = data_perm[:, :-1]
    target = data_perm[:, -1]

    # build model
    idx = 30000
    X_train, Y_train = data[:idx, :], target[:idx]
    X_test, Y_test = data[idx + 1 :, :], target[idx + 1 :]

    ordinal_features = [
        x for x in range(len(feature_names)) if x not in list(category_map.keys())
    ]
    ordinal_transformer = Pipeline(
        steps=[
            ("imputer", SimpleImputer(strategy="median")),
            ("scaler", StandardScaler()),
        ]
    )

    categorical_features = list(category_map.keys())
    categorical_transformer = Pipeline(
        steps=[
            ("imputer", SimpleImputer(strategy="median")),
            ("onehot", OneHotEncoder(handle_unknown="ignore")),
        ]
    )

    preprocessor = ColumnTransformer(
        transformers=[
            ("num", ordinal_transformer, ordinal_features),
            ("cat", categorical_transformer, categorical_features),
        ]
    )

    clf = RandomForestClassifier(n_estimators=50)

    model_pipeline = Pipeline(
        steps=[
            ("preprocess", preprocessor),
            ("classifier", clf),
        ]
    )

    model_pipeline.fit(X_train, Y_train)
    print("Model trained!")

    filename_p = "model.joblib"
    print("Saving model in %s" % filename_p)
    joblib.dump(model_pipeline, filename_p)
    print("Model saved!")


if __name__ == "__main__":
    print("Building income model...")
    main()
