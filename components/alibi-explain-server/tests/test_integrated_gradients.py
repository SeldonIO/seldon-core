import os
import tempfile

from tensorflow import keras
from tensorflow.keras.datasets import imdb
from tensorflow.keras.preprocessing import sequence

from alibiexplainer.integrated_gradients import IntegratedGradients

from .utils import download_from_gs

IMDB_KERAS_MODEL_URI = "gs://seldon-models/keras/imdb/*"
KERAS_MODEL_FILENAME = "model.h5"


def test_integrated_gradients():

    with tempfile.TemporaryDirectory() as model_dir:
        download_from_gs(IMDB_KERAS_MODEL_URI, model_dir)
        keras_model_path = os.path.join(model_dir, KERAS_MODEL_FILENAME)
        keras_model = keras.models.load_model(keras_model_path)
    integrated_gradients = IntegratedGradients(keras_model, layer=1)
    max_features = 10000
    maxlen = 100
    (x_train, y_train), (x_test, y_test) = imdb.load_data(num_words=max_features)

    x_train = sequence.pad_sequences(x_train, maxlen=maxlen)
    x_test = sequence.pad_sequences(x_test, maxlen=maxlen)

    explanation = integrated_gradients.explain(x_test[0:1].tolist())
    attrs = explanation["attributions"]
    assert len(attrs) > 0
