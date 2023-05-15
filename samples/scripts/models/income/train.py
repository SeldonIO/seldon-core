import joblib
import numpy as np
from alibi.datasets import fetch_adult
from alibi.explainers import AnchorTabular, KernelShap
from sklearn.compose import ColumnTransformer
from sklearn.ensemble import RandomForestClassifier
from sklearn.impute import SimpleImputer
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, StandardScaler


def fetch_data():
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
    X_test, Y_test = data[idx + 1:, :], target[idx + 1:]
    return X_train, Y_train, X_test, Y_test, feature_names, category_map


def train_classifier(X_train, Y_train, feature_names, category_map) -> Pipeline:
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
    return model_pipeline


def train_anchor_explainer(model_pipeline, X_train, feature_names, category_map) -> AnchorTabular:
    explainer = AnchorTabular(
        model_pipeline.predict, feature_names, categorical_names=category_map, seed=1
    )

    explainer.fit(X_train, disc_perc=(25, 50, 75))
    print("Explainer trained!")
    return explainer


def train_kernel_shap_explainer(model_pipeline, X_train, feature_names, category_map) -> KernelShap:
    explainer = KernelShap(
        model_pipeline.predict_proba,
        "identity",
        feature_names,
        categorical_names=category_map,
        seed=1
    )
    explainer.fit(X_train[:100, :])
    print("Explainer trained!")
    return explainer


if __name__ == "__main__":
    print("Building income model and explainers...")
    X_train, Y_train, X_test, Y_test, feature_names, category_map = fetch_data()
    model_pipeline = train_classifier(X_train, Y_train, feature_names, category_map)
    anchor_explainer = train_anchor_explainer(model_pipeline, X_train, feature_names, category_map)
    kernel_shap_explainer = train_kernel_shap_explainer(model_pipeline, X_train, feature_names, category_map)

    print("Saving models...")
    joblib.dump(model_pipeline, "classifier/model.joblib")
    anchor_explainer.save("explainers/anchor-explainer/data")
    kernel_shap_explainer.save("explainers/kernel-shap-explainer/data")

