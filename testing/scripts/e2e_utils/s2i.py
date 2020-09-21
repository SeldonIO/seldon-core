import logging

from sh import s2i, kind


def create_s2i_image(
    s2i_folder: str, s2i_image: str, image_name: str, s2i_runtime_image: str = None
) -> str:

    logging.info(f"Building {image_name} with s2i...")
    if s2i_runtime_image:
        s2i.build(s2i_folder, s2i_image, image_name, runtime_image=s2i_runtime_image)
    else:
        s2i.build(s2i_folder, s2i_image, image_name)

    logging.info(f"Built {image_name} with s2i")
    return image_name


def kind_load_image(image_name: str):
    logging.info(f"Loading {image_name} into Kind...")
    kind.load("docker-image", image_name)
    logging.info(f"Loaded {image_name} into Kind")
