import argparse
import os
from pathlib import Path

import tensorflow as tf
from alibi.explainers import AnchorImage


def _make_anchor_image(dirname: Path) -> None:
    url = "https://storage.googleapis.com/seldon-models/alibi-detect/classifier/"
    path_model = os.path.join(url, "cifar10", "resnet32", "model.h5")
    save_path = tf.keras.utils.get_file("resnet32", path_model)
    model = tf.keras.models.load_model(save_path)

    # we drop the first batch dimension because AnchorImage expects a single image
    image_shape = model.get_layer(index=0).input_shape[0][1:]
    alibi_model = AnchorImage(predictor=model, image_shape=image_shape)

    alibi_model.save(dirname)
    print(dirname)


def _main():
    args_parser = argparse.ArgumentParser(add_help=False)
    args_parser.add_argument(
        "--model",
        type=str,
        help="The model to create",
    )
    args_parser.add_argument(
        "--model_dir",
        type=Path,
        help="Where to save",
    )
    args = args_parser.parse_args()
    model_name = args.model
    model_dir = args.model_dir
    if model_name == "anchor_image":
        _make_anchor_image(model_dir)


if __name__ == "__main__":
    _main()
