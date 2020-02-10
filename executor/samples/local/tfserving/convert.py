import tensorflow as tf
from google.protobuf import json_format

loaded = tf.saved_model.load("./saved_model_half_plus_two_cpu/00000123")
print(list(loaded.signatures.keys()))
print(loaded)
print(loaded.signatures["serving_default"].inputs[0])
#json_string = json_format.MessageToJson(graph_def)
