import sys
from helpers import preprocess, plot
import numpy as np
import requests
from mlserver.types import InferenceRequest
from mlserver.codecs import NumpyCodec
import tensorflow_datasets as tfds
import tensorflow as tf

# Disable tensorflow debugging output
import os

os.environ["TF_CPP_MIN_LOG_LEVEL"] = "3"

# Inference variables
if len(sys.argv) < 2:
    sys.exit("Please provide inference mode (--local or --remote)")
if sys.argv[1] == "--local":
    inference_url = "http://localhost:8080/v2/models/cassava/infer"
elif sys.argv[1] == "--remote":
    inference_url = "http://localhost:8080/seldon/default/cassava/v2/models/infer"
else:
    sys.exit("Please provide inference mode (--local or --remote)")
batch_size = 16

# Load the dataset and class names
print("Lodaing dataset...")
dataset, info = tfds.load("cassava", with_info=True)
class_names = info.features["label"].names + ["unknown"]

# Shuffle the dataset with a buffer size equal to the number
# of examples in the 'validation' split
validation_dataset = dataset["validation"]
buffer_size = info.splits["validation"].num_examples
shuffled_validation_dataset = validation_dataset.shuffle(buffer_size)

# Select a batch of examples from the validation dataset
batch = (
    shuffled_validation_dataset.map(preprocess).batch(batch_size).as_numpy_iterator()
)
examples = next(batch)

# Convert the TensorFlow tensor to a numpy array
input_data = np.array(examples["image"])

# Build the inference request
inference_request = InferenceRequest(
    inputs=[NumpyCodec.encode_input(name="payload", payload=input_data)]
)

# Send the inference request and capture response
print("Sending Inference Request...")
res = requests.post(inference_url, json=inference_request.dict())
print("Got Response...")

# Parse the JSON string into a Python dictionary
response_dict = res.json()

# Extract the data array and shape from the output, assuming only
# one output or the target output is at index 0
data_list = response_dict["outputs"][0]["data"]
data_shape = response_dict["outputs"][0]["shape"]

# Convert the data list to a numpy array and reshape it
data_array = np.array(data_list).reshape(data_shape)
print("Predictions:", data_array)

# Convert the numpy array to tf tensor
data_tensor = tf.convert_to_tensor(np.squeeze(data_array), dtype=tf.int64)

# Plot the examples with their predictions
plot(examples, class_names, data_tensor)
