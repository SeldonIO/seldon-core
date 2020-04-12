from sklearn.externals import joblib
from google.protobuf.any_pb2 import Any
from iris_pb2 import IrisPredictRequest, IrisPredictResponse


class IrisClassifier:
    def __init__(self):
        self.model = joblib.load("IrisClassifier.sav")

    def predict(self, X, feature_names):
        data = parse_iris_request(X)
        score = self.model.predict_proba(data)
        return structure_iris_response(score)


def parse_iris_request(iris_request):
    unpacked = IrisPredictRequest()
    iris_request.Unpack(unpacked)
    return [
        [
            unpacked.sepal_length,
            unpacked.sepal_width,
            unpacked.petal_length,
            unpacked.petal_width,
        ]
    ]


def structure_iris_response(score):
    iris_response = IrisPredictResponse(
        setosa=score[0][0], versicolor=score[0][1], virginica=score[0][2]
    )
    response = Any()
    response.Pack(iris_response)
    return response
