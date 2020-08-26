from alibiexplainer.integrated_gradients import IntegratedGradients
import kfserving
import os
from tensorflow import keras
from tensorflow.keras.datasets import imdb
from tensorflow.keras.preprocessing import sequence
from tensorflow.keras.utils import to_categorical
IMDB_KERAS_MODEL_URI = "gs://seldon-models/keras/imdb"
KERAS_MODEL_FILENAME = "model.h5"


def test_integrated_gradients():
    os.environ.clear()
    keras_model_path = os.path.join(
        kfserving.Storage.download(IMDB_KERAS_MODEL_URI), KERAS_MODEL_FILENAME
    )
    keras_model = keras.models.load_model(keras_model_path)
    integrated_gradients = IntegratedGradients(keras_model, layer=1)
    max_features = 10000
    maxlen = 100
    (x_train, y_train), (x_test, y_test) = imdb.load_data(num_words=max_features)
    print(len(x_train), 'train sequences')
    print(len(x_test), 'test sequences')

    print('Pad sequences (samples x time)')
    x_train = sequence.pad_sequences(x_train, maxlen=maxlen)
    x_test = sequence.pad_sequences(x_test, maxlen=maxlen)
    print('x_train shape:', x_train.shape)
    print('x_test shape:', x_test.shape)

    explanation = integrated_gradients.explain(x_test[0:1].tolist())
    attrs = explanation.attributions
    print(explanation.meta)
    print('Attributions shape:', attrs.shape)
