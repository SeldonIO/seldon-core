import alibi
import dill
import numpy as np

from sklearn.preprocessing import StandardScaler, OneHotEncoder
from sklearn.impute import SimpleImputer
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer
from sklearn.ensemble import RandomForestClassifier

# make sure this directory or another, if you change, exists.
DATA_DIR = "pipeline/loanclassifier"


def load_data(train_size=30000, random_state=0):
    """Load example dataset and split between train and test datasets."""
    print("Loading adult data from alibi.")
    np.random.seed(random_state)

    data = alibi.datasets.fetch_adult()

    # mix input data
    data_perm = np.random.permutation(np.c_[data.data, data.target])
    data.data = data_perm[:, :-1]
    data.target = data_perm[:, -1]

    # perform train / test split
    X_train, y_train = data.data[:train_size, :], data.target[:train_size]
    X_test, y_test = data.data[train_size:, :], data.target[train_size:]

    return data, X_train, y_train, X_test, y_test


def train_preprocessor(data):
    """Train preprocessor."""
    print("Training preprocessor.")
    # TODO: ask if we need np.random.seed(...) here

    ordinal_features = [
        n for (n, _) in enumerate(data.feature_names)
        if n not in data.category_map
    ]

    categorical_features = list(data.category_map.keys())
    ordinal_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='median')),
            ('scaler', StandardScaler())
    ])

    categorical_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='median')),
            ('onehot', OneHotEncoder(handle_unknown='ignore'))
    ])

    preprocessor = ColumnTransformer(transformers=[
        ('num', ordinal_transformer, ordinal_features),
        ('cat', categorical_transformer, categorical_features)
    ])

    preprocessor.fit(data.data)

    return  preprocessor


def train_model(X_train, y_train, preprocessor):
    """Train model."""
    print("Training model.")
    # TODO: ask if we need np.random.seed(...) here

    clf = RandomForestClassifier(n_estimators=50)
    clf.fit(preprocessor.transform(X_train), y_train)
    return clf


def serialize_pipeline(preprocessor, clf):
    """Serialize preprocessor and model."""
    print("Serializing preprocessor and model.")

    with open(DATA_DIR + "/preprocessor.dill", "wb") as prep_f:
        dill.dump(preprocessor, prep_f)

    with open(DATA_DIR + "/model.dill", "wb") as model_f:
        dill.dump(clf, model_f)


def main():
    data, X_train, y_train, X_test, y_test = load_data()
    preprocessor = train_preprocessor(data)
    clf = train_model(X_train, y_train, preprocessor)

    serialize_pipeline(preprocessor, clf)
    return preprocessor, clf, data, X_train, y_train, X_test, y_test


if __name__ == "__main__":
    main()
