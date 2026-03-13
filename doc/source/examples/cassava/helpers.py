import numpy as np
import matplotlib.pyplot as plt
import tensorflow as tf


# Plotting function, shows the datapoints as a grid of images
# (with labels and predictions)
def plot(examples, class_names, predictions=None):
    # Label mapping
    name_map = dict(
        cmd="Mosaic Disease",
        cbb="Bacterial Blight",
        cgm="Green Mite",
        cbsd="Brown Streak Disease",
        healthy="Healthy",
        unknown="Unknown",
    )

    # Get the images, labels, and optionally predictions
    images = examples["image"]
    labels = examples["label"]
    batch_size = len(images)
    if predictions is None:
        predictions = batch_size * [None]

    # Configure the layout of the grid
    x = np.ceil(np.sqrt(batch_size))
    y = np.ceil(batch_size / x)
    fig = plt.figure(figsize=(x * 6, y * 7))

    for i, (image, label, prediction) in enumerate(zip(images, labels, predictions)):
        # Render the image
        ax = fig.add_subplot(x, y, i + 1)
        ax.imshow(image, aspect="auto")
        ax.grid(False)
        ax.set_xticks([])
        ax.set_yticks([])

        # Display the label and optionally prediction
        x_label = "Label: " + name_map[class_names[label]]
        if prediction is not None:
            x_label = (
                "Prediction: " + name_map[class_names[prediction]] + "\n" + x_label
            )
            ax.xaxis.label.set_color("green" if label == prediction else "red")
        ax.set_xlabel(x_label)

    plt.show()


# Preprocess images to be the right format for the model
def preprocess(data):
    image = data["image"]

    # Normalize [0, 255] to [0, 1]
    image = tf.cast(image, tf.float32)
    image = image / 255.0

    # Resize the images to 224 x 224
    image = tf.image.resize(image, (224, 224))

    data["image"] = image
    return data
