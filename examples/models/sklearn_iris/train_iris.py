import joblib
from sklearn.pipeline import Pipeline
from sklearn.linear_model import LogisticRegression
from sklearn import datasets


OUTPUT_FILE = "IrisClassifier.sav"


def main():
    clf = LogisticRegression(solver="liblinear", multi_class="ovr")
    p = Pipeline([("clf", clf)])
    print("Training model...")
    p.fit(X, y)
    print("Model trained!")

    print(f"Saving model in {OUTPUT_FILE}")
    joblib.dump(p, OUTPUT_FILE)
    print("Model saved!")


if __name__ == "__main__":
    print("Loading iris data set...")
    iris = datasets.load_iris()
    X, y = iris.data, iris.target
    print("Dataset loaded!")

    main()
