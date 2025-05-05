from helpers import plot, preprocess
import tensorflow as tf
import tensorflow_datasets as tfds
import tensorflow_hub as hub

# Fixes an issue with Jax and TF competing for GPU
tf.config.experimental.set_visible_devices([], "GPU")

# Load the model
model_path = "./model"
classifier = hub.KerasLayer(model_path)

# Load the dataset and store the class names
dataset, info = tfds.load("cassava", with_info=True)
class_names = info.features["label"].names + ["unknown"]

# Select a batch of examples and plot them
batch_size = 9
batch = dataset["validation"].map(preprocess).batch(batch_size).as_numpy_iterator()
examples = next(batch)
plot(examples, class_names)

# Generate predictions for the batch and plot them against their labels
predictions = classifier(examples["image"])
predictions_max = tf.argmax(predictions, axis=-1)
print(predictions_max)
plot(examples, class_names, predictions_max)
