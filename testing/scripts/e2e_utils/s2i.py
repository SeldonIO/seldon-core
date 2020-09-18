from sh import s2i, kind


def create_s2i_image(
    s2i_folder: str, s2i_image: str, image_name: str, s2i_runtime_image: str = None
) -> str:

    if s2i_runtime_image:
        s2i.build(s2i_folder, s2i_image, image_name, runtime_image=s2i_runtime_image)
    else:
        s2i.build(s2i_folder, s2i_image, image_name)

    return image_name


def kind_load_image(image_name: str):
    kind.load.docker_image(image_name)
