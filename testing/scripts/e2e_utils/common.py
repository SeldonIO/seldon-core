import os

SC_ROOT_PATH = os.path.abspath(
    os.path.join(
        __file__, os.path.pardir, os.path.pardir, os.path.pardir, os.path.pardir
    )
)
HELM_CHARTS_PATH = os.path.join(SC_ROOT_PATH, "helm-charts")


def to_helm_values_list(values):
    """
    The sh lib doesn't allow you to specify multiple instances of the same
    kwarg. https://github.com/amoffat/sh/issues/529

    The best option is to concatenate them into a list.
    """
    values_list = []
    for key, val in values.items():
        values_list += ["--set", f"{key}={val}"]

    return values_list
