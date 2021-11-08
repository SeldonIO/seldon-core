
import numpy as np
np.random.seed(123)   # for reproducibility

from tensorflow.keras.layers import (Activation, Convolution2D, Dense, Dropout,
                                     Flatten, MaxPooling2D)
from tensorflow.keras.models import Sequential
from tensorflow.keras.utils import to_categorical

from tensorflow.keras.datasets import mnist


# Load pre-shuffled MNIST data into train and test sets
(X_train, y_train), (X_test, y_test) = mnist.load_data()

# Reshape data for Tensorflow
X_train = X_train.reshape(-1, 28, 28, 1)
X_test = X_test.reshape(-1, 28, 28, 1)

X_train = X_train.astype("float32")
X_test = X_test.astype("float32")
X_train /= 255
X_test /= 255

# Convert 1-dimensional class arrays to 10-dimensional class matrices
Y_train = to_categorical(y_train, 10)
Y_test = to_categorical(y_test, 10)

# define model
model = Sequential()

# declare input layer
model.add(Convolution2D(32, (3, 3), activation="relu", input_shape=(28, 28, 1)))

# add more layers
model.add(Convolution2D(32, (3, 3), activation="relu"))
model.add(MaxPooling2D(pool_size=(2, 2)))
model.add(Dropout(0.25))

# and even more layers
model.add(Flatten())
model.add(Dense(128, activation="relu"))
model.add(Dropout(0.5))
model.add(Dense(10, activation="softmax"))

model.summary()

# compile model and train
model.compile(loss="categorical_crossentropy", optimizer="adam", metrics=["accuracy"])
model.fit(X_train, Y_train, epochs=4, batch_size=32, verbose=1)

score = model.evaluate(X_test, Y_test, verbose=0)
print("\n%s: %.2f%%" % (model.metrics_names[1], score[1] * 100))

# save model
model.save("models-repository/mnist/1/model.savedmodel")
